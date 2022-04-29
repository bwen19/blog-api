package api

import (
	mockdb "blog/db/mock"
	db "blog/db/sqlc"
	"blog/token"
	"blog/util"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomComment(t *testing.T, commenter db.User, articleID int64) db.Comment {
	return db.Comment{
		ID:        util.RandomInt(1, 1000),
		ParentID:  sql.NullInt64{Int64: 0, Valid: false},
		ArticleID: articleID,
		Commenter: commenter.Username,
		Content:   util.RandomString(20),
	}
}

func TestListCommentsAPI(t *testing.T) {
	commenter, _ := randomUser(t)
	article := randomArticle(t, commenter)
	comment := randomComment(t, commenter, article.ID)
	cL := commentList{
		ID:        comment.ID,
		ParentID:  comment.ParentID,
		ArticleID: comment.ArticleID,
		Commenter: comment.Commenter,
		AvatarSrc: commenter.AvatarSrc,
		Content:   comment.Content,
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: fmt.Sprintf("page_id=1&page_size=5&article_id=%d", comment.ArticleID),
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListCommentsByArticleParams{
					Limit:     5,
					Offset:    0,
					ArticleID: article.ID,
				}
				store.EXPECT().
					ListCommentsByArticle(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return([]commentList{cL}, nil)
				store.EXPECT().
					ListChildComments(gomock.Any(), gomock.Eq([]int64{comment.ID})).
					Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "InternalError1",
			query: fmt.Sprintf("page_id=1&page_size=5&article_id=%d", comment.ArticleID),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCommentsByArticle(gomock.Any(), gomock.Any()).
					Times(1).Return([]commentList{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InternalError2",
			query: fmt.Sprintf("page_id=1&page_size=5&article_id=%d", comment.ArticleID),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCommentsByArticle(gomock.Any(), gomock.Any()).
					Times(1).Return([]commentList{cL}, nil)
				store.EXPECT().
					ListChildComments(gomock.Any(), gomock.Any()).
					Times(1).Return([]db.ListChildCommentsRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidID",
			query: fmt.Sprintf("page_id=1&page_size=5&article_id=%d", 0),
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/comments?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestCreateComment(t *testing.T) {
	commenter, _ := randomUser(t)
	article := randomArticle(t, commenter)
	comment := randomComment(t, commenter, article.ID)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"article_id": article.ID,
				"content":    comment.Content,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, commenter.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: comment.Commenter,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(commenter, nil)
				arg2 := db.CreateCommentParams{
					ArticleID: comment.ArticleID,
					Commenter: comment.Commenter,
					Content:   comment.Content,
				}
				store.EXPECT().
					CreateComment(gomock.Any(), gomock.Eq(arg2)).
					Times(1).Return(comment, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidArticleID",
			body: gin.H{
				"article_id": 0,
				"content":    comment.Content,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, commenter.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: comment.Commenter,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(commenter, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/user/comments"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteComment(t *testing.T) {
	commenter, _ := randomUser(t)
	article := randomArticle(t, commenter)
	comment := randomComment(t, commenter, article.ID)

	testCases := []struct {
		name          string
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, commenter.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: comment.Commenter,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(commenter, nil)
				arg2 := db.DeleteCommentParams{
					ID:        comment.ID,
					Commenter: comment.Commenter,
				}
				store.EXPECT().
					DeleteComment(gomock.Any(), gomock.Eq(arg2)).
					Times(1).Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/user/comments/%d", comment.ID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
