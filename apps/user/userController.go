package user

import (
	"strconv"
	"encoding/json"
	"net/http"
	
	"linq/core/api"	    
	"linq/core/utils"	    
	. "linq/core/repository"
	
	"github.com/gorilla/mux"
)

type userController struct{
	repo IRepository
}

func NewUserController(repo IRepository) userController{
	return userController{
		repo: repo,
	}
}

func (ctrl userController)GetAll(w http.ResponseWriter, r *http.Request) {
	users := ctrl.repo.GetAll()
	response := api.JsonDTResponse{
		Draw: r.URL.Query().Get("draw"),
		RecordsTotal: ctrl.repo.CountAll(),
		RecordsFiltered: len(users),
		Data: users,
		Success: true,
	}
	
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
        utils.HandleWarn(err)
	}
}

func (ctrl userController)Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
    userId, err := strconv.Atoi(vars["id"])
    utils.HandleWarn(err)
    
	user := ctrl.repo.Get(userId)
	
	response := api.NewJsonResponse(user, (user.GetId() > 0), 1, 1, "")
	
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HandleWarn(err)
	}
}