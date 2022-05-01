package api

import (
	mockdb "blog/server/db/mock"
	db "blog/server/db/sqlc"
	"blog/server/token"
	"blog/server/util"
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
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func randomTag(t *testing.T) db.Tag {
	return db.Tag{
		Name:  util.RandomString(6),
		Count: 0,
	}
}

func requireBodyMatchTag(t *testing.T, body *bytes.Buffer, tag db.Tag) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotTag db.Tag
	err = json.Unmarshal(data, &gotTag)
	require.NoError(t, err)
	require.Equal(t, gotTag.Name, tag.Name)
	require.Equal(t, gotTag.Count, tag.Count)
}

func TestCreateTag(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	tag := randomTag(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name": tag.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetUserParams{
					Username: admin.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return(admin, nil)
				store.EXPECT().
					CreateTag(gomock.Any(), gomock.Eq(tag.Name)).
					Times(1).Return(tag, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTag(t, recorder.Body, tag)
			},
		},
		{
			name: "InvalidName",
			body: gin.H{
				"name": "",
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateName",
			body: gin.H{
				"name": tag.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					CreateTag(gomock.Any(), gomock.Eq(tag.Name)).
					Times(1).Return(db.Tag{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"name": tag.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					CreateTag(gomock.Any(), gomock.Eq(tag.Name)).
					Times(1).Return(db.Tag{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/api/admin/tags"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetTag(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	tag := randomTag(t)

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetUserParams{
					Username: admin.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return(admin, nil)
				store.EXPECT().
					GetTag(gomock.Any(), gomock.Eq(tag.Name)).
					Times(1).Return(tag, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTag(t, recorder.Body, tag)
			},
		},
		{
			name: "NotFound",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					GetTag(gomock.Any(), gomock.Eq(tag.Name)).
					Times(1).Return(db.Tag{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					GetTag(gomock.Any(), gomock.Eq(tag.Name)).
					Times(1).Return(db.Tag{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/api/admin/tags/%s", tag.Name)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListTags(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	tag := randomTag(t)

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "page_id=1&page_size=5",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: admin.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(admin, nil)
				arg2 := db.ListTagsParams{
					Limit:  5,
					Offset: 0,
				}
				store.EXPECT().
					ListTags(gomock.Any(), gomock.Eq(arg2)).
					Times(1).Return([]db.Tag{tag}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "InvalidPageSize",
			query: "page_id=1&page_size=4",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "InternalError",
			query: "page_id=1&page_size=5",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					ListTags(gomock.Any(), gomock.Any()).
					Times(1).Return([]db.Tag{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/api/admin/tags?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateTag(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	tag := randomTag(t)
	newTag := tag
	newTag.Name = "someNewName"
	newTag.Count = 10

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"new_name": newTag.Name,
				"count":    newTag.Count,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.GetUserParams{
					Username: admin.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg1)).
					Times(1).Return(admin, nil)
				arg2 := db.UpdateTagParams{
					Name:       tag.Name,
					SetNewName: true,
					NewName:    newTag.Name,
					SetCount:   true,
					Count:      newTag.Count,
				}
				store.EXPECT().
					UpdateTag(gomock.Any(), gomock.Eq(arg2)).
					Times(1).Return(newTag, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTag(t, recorder.Body, newTag)
			},
		},
		{
			name: "InvalidCount",
			body: gin.H{
				"new_name": newTag.Name,
				"count":    -1,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateName",
			body: gin.H{
				"new_name": newTag.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					UpdateTag(gomock.Any(), gomock.Any()).
					Times(1).Return(db.Tag{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"new_name": newTag.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					UpdateTag(gomock.Any(), gomock.Any()).
					Times(1).Return(db.Tag{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/api/admin/tags/%s", tag.Name)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteTag(t *testing.T) {
	admin := db.User{
		Username: "admin",
		Role:     "admin",
	}
	tag := randomTag(t)

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetUserParams{
					Username: admin.Username,
				}
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return(admin, nil)
				store.EXPECT().
					DeleteTag(gomock.Any(), gomock.Any()).
					Times(1).Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					DeleteTag(gomock.Any(), gomock.Any()).
					Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/api/admin/tags/%s", tag.Name)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
