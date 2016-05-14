package controllers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"bitbucket.org/sofyan_a/linq.im/core/api"
	"bitbucket.org/sofyan_a/linq.im/core/services"
	"bitbucket.org/sofyan_a/linq.im/core/utils"
	"bitbucket.org/sofyan_a/linq.im/domains/users"

	"github.com/satori/go.uuid"
)

type authController struct {
	service services.IService
}

type RequestAuthDataModel struct {
	Data  users.User `json:"data"`
	Token string     `json:"token"`
}

type RequestDataAuthCredential struct {
	Data  users.UserCredential `json:"data"`
	Token string               `json:"token"`
}

func AuthController(service services.IService) authController {
	return authController{
		service: service,
	}
}


func (ctrl authController) GetAll(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	length, err := strconv.Atoi(respWriter.FormValue("length"))
	if err != nil {
		length = 25
		err = nil
	}

	search := respWriter.FormValue("search[value]")

	orderBy, err := strconv.Atoi(respWriter.FormValue("order[0][column]"))
	if err != nil {
		orderBy = 1
		err = nil
	}

	orderDir := respWriter.FormValue("order[0][dir]")

	draw, err := strconv.Atoi(respWriter.FormValue("draw"))
	if err != nil {
		draw = 0
		err = nil
	}

	paging := utils.Paging{
		search,
		length,
		orderBy,
		orderDir,
	}

	users, err := ctrl.service.GetAll(paging)
	respWriter.HandleApiError(err, http.StatusInternalServerError)

	count, err := ctrl.service.CountAll()
	respWriter.HandleApiError(err, http.StatusInternalServerError)

	respWriter.DTJsonResponse(users, (users != nil), count, users.GetLength(), draw)
}

func (ctrl authController) Get(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	userId, err := uuid.FromString(respWriter.MuxVars("id"))
	respWriter.HandleApiError(err, http.StatusBadRequest)

	user, err := ctrl.service.Get(userId)
	respWriter.HandleApiError(err, http.StatusInternalServerError)

	respWriter.ReturnJson(user)
}

func (ctrl authController) Create(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	var requestData RequestAuthDataModel
	err := respWriter.DecodeBody(&requestData)

	if err == nil {
		err = ctrl.service.Create(&requestData.Data)
		respWriter.HandleApiError(err, http.StatusInternalServerError)

		if err == nil {
			respWriter.ReturnJson(requestData.Data)
		}
	}
}

func (ctrl authController) Modify(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	userId, err := uuid.FromString(respWriter.MuxVars("id"))
	respWriter.HandleApiError(err, http.StatusBadRequest)

	if err == nil {
		var requestData RequestAuthDataModel
		err = respWriter.DecodeBody(&requestData)
		respWriter.HandleApiError(err, http.StatusBadRequest)

		if err == nil {
			requestData.Data.Uid = userId

			err = ctrl.service.Modify(&requestData.Data)
			if err == nil {
				respWriter.HandleApiError(err, http.StatusInternalServerError)
				respWriter.ReturnJson(requestData.Data)
			}
		}
	}
}

func (ctrl authController) SetUserPhoto(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	userId, err := uuid.FromString(respWriter.MuxVars("id"))
	respWriter.HandleApiError(err, http.StatusBadRequest)

	if err == nil {
		var requestData api.RequestDataImage

		respWriter.DecodeBody(&requestData)

		plainBase64 := strings.Replace(requestData.Data, "data:image/png;base64,", "", 1)

		imageReader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(plainBase64))

		fileName := fmt.Sprintf("%s.png", userId)

		img, err := os.Create("./uploads/user_avatars/" + fileName)
		respWriter.HandleApiError(err, http.StatusInternalServerError)

		if err == nil {
			defer img.Close()
			_, err = io.Copy(img, imageReader)
			respWriter.HandleApiError(err, http.StatusInternalServerError)

			if err == nil {
				userModel, err := ctrl.service.Get(userId)
				respWriter.HandleApiError(err, http.StatusInternalServerError)

				if err == nil {
					user := userModel.(*users.User)
					user.Avatar.String = fileName
					user.Avatar.Valid = true

					userService := ctrl.service.(users.UserService)
					err = userService.UpdateUserPhoto(user)
					respWriter.HandleApiError(err, http.StatusInternalServerError)

					if err == nil {
						respWriter.ReturnJson(user)
					}
				}
			}
		}
	}
}

func (ctrl authController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	userId, err := uuid.FromString(respWriter.MuxVars("id"))
	respWriter.HandleApiError(err, http.StatusBadRequest)

	var requestData RequestDataAuthCredential

	err = respWriter.DecodeBody(&requestData)
	respWriter.HandleApiError(err, http.StatusBadRequest)

	if err == nil {
		requestData.Data.Uid = userId

		userService := ctrl.service.(users.UserService)

		err := userService.ChangePassword(&requestData.Data)
		respWriter.HandleApiError(err, http.StatusBadRequest)

		if err == nil {
			respWriter.ReturnJson(requestData.Data)
		}
	}
}

func (ctrl authController) Remove(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	userId, err := uuid.FromString(respWriter.MuxVars("id"))
	respWriter.HandleApiError(err, http.StatusBadRequest)

	if err == nil {
		if exist, err := ctrl.service.IsExist(userId); !exist {
			respWriter.HandleApiError(err, http.StatusBadRequest)
		}
		user, err := ctrl.service.Get(userId)
		respWriter.HandleApiError(err, http.StatusBadRequest)

		err = ctrl.service.Remove(user)
		respWriter.HandleApiError(err, http.StatusInternalServerError)

		respWriter.ReturnJson(user)
	}
}

func (ctrl authController) RemoveBulk(w http.ResponseWriter, r *http.Request) {
	respWriter := api.ApiService(w, r)

	var requestData api.RequestDataIds

	respWriter.DecodeBody(&requestData)

	result := ctrl.service.RemoveBulk(requestData.Data.Ids)

	respWriter.ReturnJson(result)
}
