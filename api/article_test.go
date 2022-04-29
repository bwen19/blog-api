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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomArticle(t *testing.T, author db.User) db.Article {
	return db.Article{
		ID:        util.RandomInt(1, 1000),
		Author:    author.Username,
		Category:  util.RandomString(6),
		Title:     util.RandomString(10),
		Summary:   util.RandomString(20),
		Content:   util.RandomString(100),
		Status:    "published",
		ViewCount: 0,
	}
}

func newArticleListRow(t *testing.T, article db.Article) db.ListArticlesRow {
	return db.ListArticlesRow{
		ID:        article.ID,
		Author:    article.Author,
		Category:  article.Category,
		Title:     article.Title,
		Summary:   article.Summary,
		Status:    article.Status,
		ViewCount: article.ViewCount,
		UpdateAt:  article.UpdateAt,
	}
}

func requireBodyMatchArticle(t *testing.T, body *bytes.Buffer, article db.Article, author db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotArticle articleResponse
	err = json.Unmarshal(data, &gotArticle)
	require.NoError(t, err)
	require.Equal(t, gotArticle.ID, article.ID)
	require.Equal(t, gotArticle.Author, author.Username)
	require.Equal(t, gotArticle.AvatarSrc, author.AvatarSrc)
	require.Equal(t, gotArticle.Title, article.Title)
	require.Equal(t, gotArticle.Summary, article.Summary)
	require.Equal(t, gotArticle.Content, article.Content)
	require.Equal(t, gotArticle.ViewCount, article.ViewCount)
	require.WithinDuration(t, gotArticle.UpdateAt, article.UpdateAt, time.Second)
}

func TestReadPublishedArticleAPI(t *testing.T) {
	author, _ := randomUser(t)
	article := randomArticle(t, author)

	testCases := []struct {
		name          string
		uri           int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  article.ID,
			buildStubs: func(store *mockdb.MockStore) {
				articleRsp := article
				articleRsp.ViewCount++
				store.EXPECT().
					ReadArticle(gomock.Any(), gomock.Eq(article.ID)).
					Times(1).Return(articleRsp, nil)
				arg := db.GetUserParams{
					Username: article.Author,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return(author, nil)
				store.EXPECT().
					ListArticleTags(gomock.Any(), gomock.Eq(article.ID)).
					Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				article.ViewCount++
				requireBodyMatchArticle(t, recorder.Body, article, author)
			},
		},
		{
			name: "InternalError",
			uri:  article.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ReadArticle(gomock.Any(), gomock.Any()).
					Times(1).Return(db.Article{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			uri:  0,
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

			url := fmt.Sprintf("/api/articles/%d", tc.uri)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListPublishedArticles(t *testing.T) {
	author, _ := randomUser(t)
	articles := []db.ListArticlesRow{}
	for i := 0; i < 10; i++ {
		newArticle := randomArticle(t, author)
		articles = append(articles, newArticleListRow(t, newArticle))
	}

	testCases := []struct {
		name          string
		param         listPublishedArticlesRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			param: listPublishedArticlesRequest{
				SortBy:   "time",
				PageID:   1,
				PageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListArticlesParams{
					Limit:       5,
					Offset:      0,
					Status:      "published",
					AnyAuthor:   true,
					AnyCategory: true,
					AnyTag:      true,
					TimeDesc:    true,
					CountDesc:   false,
				}
				store.EXPECT().
					ListArticles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(articles, nil)
				store.EXPECT().
					ListArticleTags(gomock.Any(), gomock.Any()).
					Times(10)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
			},
		},
		{
			name: "ListByCategory",
			param: listPublishedArticlesRequest{
				SortBy:   "time",
				PageID:   1,
				PageSize: 5,
				Category: "cate",
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListArticlesParams{
					Limit:       5,
					Offset:      0,
					Status:      "published",
					AnyAuthor:   true,
					AnyCategory: false,
					Category:    "cate",
					AnyTag:      true,
					TimeDesc:    true,
					CountDesc:   false,
				}
				store.EXPECT().
					ListArticles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(articles, nil)
				store.EXPECT().
					ListArticleTags(gomock.Any(), gomock.Any()).
					Times(10)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
			},
		},
		{
			name: "InvalidSort",
			param: listPublishedArticlesRequest{
				SortBy:   "invalid",
				PageID:   1,
				PageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
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

			query := fmt.Sprintf("sort_by=%s&page_id=%d&page_size=%d", tc.param.SortBy, tc.param.PageID, tc.param.PageSize)
			if tc.param.Author != "" {
				query = fmt.Sprintf("%s&author=%s", query, tc.param.Author)
			}
			if tc.param.Category != "" {
				query = fmt.Sprintf("%s&category=%s", query, tc.param.Category)
			}
			if tc.param.Tag != "" {
				query = fmt.Sprintf("%s&tag=%s", query, tc.param.Tag)
			}
			url := fmt.Sprintf("/api/articles?%s", query)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestCreateArticleByAuthorAPI(t *testing.T) {
	author, _ := randomUser(t)
	author.Role = "author"
	category := randomCategory(t)
	tag := randomTag(t)
	article := randomArticle(t, author)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"title":   article.Title,
				"summary": article.Summary,
				"content": article.Content,
				// "category": category.Name,
				"tags": []string{"aa", "bb", "cc"},
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, author.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: author.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(author, nil)
				store.EXPECT().
					GetCategory(gomock.Any(), gomock.Eq("default")).
					Times(1).Return("default", sql.ErrNoRows)
				store.EXPECT().
					CreateCategory(gomock.Any(), gomock.Any()).
					Times(1).Return("default", nil)
				arg2 := db.CreateArticleParams{
					Author:   author.Username,
					Category: "default",
					Title:    article.Title,
					Summary:  article.Summary,
					Content:  article.Content,
				}

				store.EXPECT().
					CreateArticle(gomock.Any(), gomock.Eq(arg2)).
					Times(1).Return(article, nil)

				store.EXPECT().
					ListArticleTags(gomock.Any(), gomock.Eq(article.ID)).
					Times(1).Return([]string{"cc", "dd"}, nil)
				store.EXPECT().
					GetTag(gomock.Any(), gomock.Any()).
					Times(2).Return(tag, sql.ErrNoRows)
				store.EXPECT().
					CreateTag(gomock.Any(), gomock.Any()).
					Times(2).Return(tag, nil)
				store.EXPECT().
					CreateArticleTag(gomock.Any(), gomock.Any()).
					Times(2).Return(db.ArticleTag{}, nil)
				store.EXPECT().
					DeleteArticleTag(gomock.Any(), gomock.Any()).
					Times(1).Return(nil)
				store.EXPECT().
					UpdateTag(gomock.Any(), gomock.Any()).
					Times(3).Return(db.Tag{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticle(t, recorder.Body, article, author)
			},
		},
		{
			name: "TooManyTags",
			body: gin.H{
				"title":    article.Title,
				"summary":  article.Summary,
				"content":  article.Content,
				"category": category.Name,
				"tags":     []string{"aa", "bb", "cc", "dd", "ee", "ff"},
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, author.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: author.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(author, nil)
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

			url := "/api/author/articles"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
