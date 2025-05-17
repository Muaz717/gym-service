package personHandler_test

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/Muaz717/github.com/Muaz717/gym_app/app/internal/http/handlers/person"
	"github.com/Muaz717/github.com/Muaz717/gym_app/app/internal/http/handlers/person/mocks"
	"github.com/Muaz717/github.com/Muaz717/gym_app/app/internal/lib/logger/handlers/slogdiscard"
	"github.com/Muaz717/github.com/Muaz717/gym_app/app/internal/models"
	"github.com/Muaz717/github.com/Muaz717/gym_app/app/internal/services/person"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddPerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type mockBehavior func(s *mocks.PersonService, person models.Person)

	tests := []struct {
		name         string
		inputBody    string
		inputPerson  models.Person
		mockBehavior mockBehavior
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			inputBody: `{
				"name": "John Doe",
				"phone": "12345678910"
			}`,
			inputPerson: models.Person{Name: "John Doe", Phone: "12345678910"},
			mockBehavior: func(s *mocks.PersonService, person models.Person) {
				s.On("AddPerson", mock.Anything, person).Return(1, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"OK","msg":"Person added, personId: 1"}`,
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{}`,
			inputPerson:  models.Person{},
			mockBehavior: func(s *mocks.PersonService, person models.Person) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"Name":"ФИО обязательно для заполнения", "Phone":"Телефон обязателен для заполнения"}`,
		},
		{
			name: "Person already exists",
			inputBody: `{
				"name": "John Doe",
				"phone": "12345678910"
			}`,
			inputPerson: models.Person{Name: "John Doe", Phone: "12345678910"},
			mockBehavior: func(s *mocks.PersonService, person models.Person) {
				s.On("AddPerson", mock.Anything, person).Return(0, personService.ErrPersonExists)
			},
			expectedCode: http.StatusConflict,
			expectedBody: `{"status":"Error","error":"person already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init mocks
			serviceMock := mocks.NewPersonService(t)
			tt.mockBehavior(serviceMock, tt.inputPerson)

			// init handler
			handler := personHandler.personHandler.New(context.Background(), slogdiscard.NewDiscardLogger(), serviceMock)

			// init router
			r := gin.New()
			r.POST("/people/add", handler.AddPerson)

			// perform request
			req := httptest.NewRequest(http.MethodPost, "/people/add", bytes.NewBufferString(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			require.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestUpdatePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type mockBehavior func(s *mocks.PersonService, person models.Person, id int)

	tests := []struct {
		name         string
		personID     string
		inputBody    string
		inputPerson  models.Person
		mockBehavior mockBehavior
		expectedCode int
		expectedBody string
	}{
		{
			name:     "Success",
			personID: "1",
			inputBody: `{
				"name": "Updated Name",
				"phone": "12345678910"
			}`,
			inputPerson: models.Person{Name: "Updated Name", Phone: "12345678910"},
			mockBehavior: func(s *mocks.PersonService, person models.Person, id int) {
				s.On("UpdatePerson", mock.Anything, person, id).Return(1, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"OK","msg":"Person updated"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewPersonService(t)
			tt.mockBehavior(serviceMock, tt.inputPerson, 1)

			handler := personHandler.New(context.Background(), slogdiscard.NewDiscardLogger(), serviceMock)

			r := gin.New()
			r.PUT("/people/update/:id", handler.UpdatePerson)

			req := httptest.NewRequest(http.MethodPut, "/people/update/1", bytes.NewBufferString(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			require.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestDeletePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type mockBehavior func(s *mocks.PersonService, id int)

	tests := []struct {
		name         string
		personID     string
		mockBehavior mockBehavior
		expectedCode int
		expectedBody string
	}{
		{
			name:     "Success",
			personID: "1",
			mockBehavior: func(s *mocks.PersonService, id int) {
				s.On("DeletePerson", mock.Anything, id).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"OK","msg":"Person deleted"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewPersonService(t)
			tt.mockBehavior(serviceMock, 1)

			handler := personHandler.New(context.Background(), slogdiscard.NewDiscardLogger(), serviceMock)

			r := gin.New()
			r.DELETE("/people/delete/:id", handler.DeletePerson)

			req := httptest.NewRequest(http.MethodDelete, "/people/delete/1", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			require.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestFindPersonByName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type mockBehavior func(s *mocks.PersonService, name string)

	tests := []struct {
		name         string
		queryName    string
		mockBehavior mockBehavior
		expectedCode int
		expectedBody string
	}{
		{
			name:      "Success",
			queryName: "John",
			mockBehavior: func(s *mocks.PersonService, name string) {
				s.On("FindPersonByName", mock.Anything, name).Return(models.Person{
					Id: 1, Name: "John", Phone: "12345678910",
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":1,"name":"John","phone":"12345678910"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewPersonService(t)
			tt.mockBehavior(serviceMock, tt.queryName)

			handler := personHandler.New(context.Background(), slogdiscard.NewDiscardLogger(), serviceMock)

			r := gin.New()
			r.GET("/people/find", handler.FindPersonByName)

			req := httptest.NewRequest(http.MethodGet, "/people/find?name="+tt.queryName, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			require.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestFindAllPeople(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockBehavior func(s *mocks.PersonService)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			mockBehavior: func(s *mocks.PersonService) {
				s.On("FindAllPeople", mock.Anything).Return([]models.Person{
					{Id: 1, Name: "John", Phone: "12345678910"},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `[{"id":1,"name":"John","phone":"12345678910"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewPersonService(t)
			tt.mockBehavior(serviceMock)

			handler := personHandler.New(context.Background(), slogdiscard.NewDiscardLogger(), serviceMock)

			r := gin.New()
			r.GET("/people", handler.FindAllPeople)

			req := httptest.NewRequest(http.MethodGet, "/people", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			require.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
