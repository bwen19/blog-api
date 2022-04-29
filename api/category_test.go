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
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func randomCategory(t *testing.T) db.Category {
	return db.Category{
		Name: util.RandomString(6),
	}
}

type categoryResponse struct {
	Category string `json:"category"`
}

func requireBodyMatchCategory(t *testing.T, body *bytes.Buffer, category db.Category) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotCategory categoryResponse
	err = json.Unmarshal(data, &gotCategory)
	require.NoError(t, err)
	require.Equal(t, gotCategory.Category, category.Name)
}

func TestCreateCategory(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	category := randomCategory(t)

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
				"name": category.Name,
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
					CreateCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return(category.Name, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCategory(t, recorder.Body, category)
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
				"name": category.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					CreateCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return("", &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"name": category.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					CreateCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return("", sql.ErrConnDone)
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

			url := "/api/admin/categories"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetCategory(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	category := randomCategory(t)

	testCases := []struct {
		name          string
		uri           string
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  category.Name,
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
					GetCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return(category.Name, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCategory(t, recorder.Body, category)
			},
		},
		{
			name: "NotFound",
			uri:  category.Name,
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					GetCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return("", sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			uri:  category.Name,
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					GetCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return("", sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/admin/categories/%s", tc.uri)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListCategories(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	category := randomCategory(t)

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
					ListCategories(gomock.Any()).
					Times(1).Return([]string{category.Name}, nil)
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
					ListCategories(gomock.Any()).
					Times(1).Return([]string{}, sql.ErrConnDone)
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

			url := "/api/admin/categories"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	category := randomCategory(t)
	newCategory := category
	newCategory.Name = "someNewName"

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
				"new_name": newCategory.Name,
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
				arg2 := db.UpdateCategoryParams{
					Name:    category.Name,
					NewName: newCategory.Name,
				}
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Eq(arg2)).
					Times(1).Return(newCategory.Name, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCategory(t, recorder.Body, newCategory)
			},
		},
		{
			name: "InvalidNewName",
			body: gin.H{
				"new_name": "",
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
			name: "NoNeedUpdate",
			body: gin.H{
				"new_name": category.Name,
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
				"new_name": newCategory.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Any()).
					Times(1).Return("", &pq.Error{Code: "23505"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"new_name": newCategory.Name,
			},
			setupAuth: func(request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(admin, nil)
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Any()).
					Times(1).Return("", sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/admin/categories/%s", category.Name)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	admin := db.User{
		Username: "su",
		Role:     "admin",
	}
	category := randomCategory(t)

	testCases := []struct {
		name          string
		uri           string
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  category.Name,
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
					DeleteCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			uri:  category.Name,
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
					DeleteCategory(gomock.Any(), gomock.Eq(category.Name)).
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

			url := fmt.Sprintf("/api/admin/categories/%s", tc.uri)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
