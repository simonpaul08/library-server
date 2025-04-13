package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	"project/libraryManagement/config"
	"project/libraryManagement/models"
	"project/libraryManagement/utils"
	"time"
)

type CreateLibrary struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type InventoryStruct struct {
	Title       string         `json:"title"`
	Authors     pq.StringArray `json:"authors"`
	Publisher   string         `json:"publisher"`
	Version     string         `json:"version"`
	TotalCopies uint           `json:"totalCopies"`
}
type AddBookStruct struct {
	ISBN   uint `json:"isbn"`
	Copies uint `json:"copies"`
}
type UpdateBookStruct struct {
	ISBN        uint           `json:"isbn"`
	Title       string         `json:"title"`
	Authors     pq.StringArray `json:"authors"`
	Publisher   string         `json:"publisher"`
	Version     string         `json:"version"`
	TotalCopies uint           `json:"totalCopies"`
}
type SearchBookStruct struct {
	Query string `json:"query"`
}
type IssueBookStruct struct {
	ISBN uint `json:"isbn"`
}
type ApproveRequestStruct struct {
	ReqId uint `json:"reqId"`
}

// register a new library and adding a new user as owner
func RegisterLibrary(c *gin.Context) {
	var library CreateLibrary

	err := c.ShouldBind(&library)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check duplicate - library
	lib, _ := utils.FindLibrary(library.Name)
	if lib != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "Library with this name already exists"})
		return
	}

	// check duplicate - user
	duplicateUser, _ := utils.FindUser(library.Email)
	if duplicateUser != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "Email already in use"})
		return
	}

	newLib := models.Library{Name: library.Name}

	newLibRes := config.DB.Create(&newLib)
	if newLibRes.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": newLibRes.Error.Error()})
		return
	}

	newUser := models.Users{Email: library.Email, LibID: newLib.ID, Role: "owner"}

	newUserRes := config.DB.Create(&newUser)
	if newUserRes.Error != nil {
		// delete the library as well
		config.DB.Delete(&models.Library{}, newLib.ID)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": newLibRes.Error.Error()})
		return
	}

	// send otp for verification
	result := utils.SendOTP(library.Email)
	if result != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error sending otp"})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "otp sent successfully"})
}

// Create BookInventory
func CreateInventory(c *gin.Context) {
	var data InventoryStruct
	var Inventory models.BookInventory

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	owner, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// if book is not present in lib - create
	// if book is present in lib - +1

	// check if book is present in library
	res := config.DB.Preload("Library").Where("title = ? AND lib_id = ?", data.Title, owner.LibID).First(&Inventory)

	if res.Error == nil {
		// inventory exists
		Inventory.AvailableCopies += data.TotalCopies
		Inventory.TotalCopies += data.TotalCopies
		res := config.DB.Save(&Inventory)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "unable to add book"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "book added to the inventory"})
	} else {
		// gerenrate qr code
		qr, err := utils.GenerateQR(data.Title)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error generating qr code"})
			return
		}
		// inventory doesn't exists
		item := models.BookInventory{Title: data.Title, Authors: data.Authors, Publisher: data.Publisher, Version: data.Version, TotalCopies: data.TotalCopies, AvailableCopies: data.TotalCopies, LibID: owner.LibID, QrCode: qr}
		res := config.DB.Create(&item)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "unable to create inventory", "err": res.Error.Error()})
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"message": "inventory created successfully"})
	}
}

// Remove a book
func RemoveBook(c *gin.Context) {
	id := c.Param("id")
	var Inventory models.BookInventory

	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "isbn is required"})
		return
	}

	// check if inventory is present in the library
	book := config.DB.Where("isbn = ?", id).First(&Inventory)
	if book.Error != nil {
		// doesn't exists
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book inventory does not exists"})
		return
	}

	if Inventory.TotalCopies > 1 && Inventory.AvailableCopies > 1 {
		Inventory.TotalCopies -= 1
		Inventory.AvailableCopies -= 1
		res := config.DB.Save(&Inventory)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "unable to remove book"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "book removed successfully"})
	} else {
		if Inventory.TotalCopies > 1 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "issued books cannot be removed"})
		} else {
			// remove inventory
			del := config.DB.Where("isbn = ?", id).Delete(&Inventory)
			if del.Error != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "unable to remove book"})
			}
			c.IndentedJSON(http.StatusOK, gin.H{"message": "inventory removed successfully"})
		}
	}
}

// Add book
func AddBook(c *gin.Context) {
	var data AddBookStruct
	var Inventory models.BookInventory

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if inventory exists
	isInventory := config.DB.Where("isbn = ?", data.ISBN).First(&Inventory)
	if isInventory.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "inventory does not exists"})
		return
	}

	Inventory.TotalCopies += data.Copies
	Inventory.AvailableCopies += data.Copies

	// save
	update := config.DB.Save(&Inventory)
	if update.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error updating book inventory"})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "book(s) added successfully"})
}

// Update book
func UpdateBook(c *gin.Context) {
	var data UpdateBookStruct
	var Inventory models.BookInventory

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if inventory exists
	isInventory := config.DB.Where("isbn = ?", data.ISBN).First(&Inventory)
	if isInventory.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "inventory does not exists"})
		return
	}

	// generate qr code
	qr, e := utils.GenerateQR(data.Title)
	if e != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error generating qr"})
		return
	}

	// update the inventory
	update := config.DB.Where("isbn = ?", data.ISBN).Updates(models.BookInventory{Title: data.Title, Authors: data.Authors, Publisher: data.Publisher,
		Version: data.Version, QrCode: qr, TotalCopies: data.TotalCopies + Inventory.TotalCopies, AvailableCopies: data.TotalCopies + Inventory.AvailableCopies,})
	if update.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error updating the inventory"})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Inventory updated successfully"})
}

// retrieve books by libid
func RetrieveBooksByLib(c *gin.Context) {
	var Library models.Library
	var Books []models.BookInventory

	id := c.Param("id")
	if id == "" {
		c.IndentedJSON(http.StatusBadGateway, gin.H{"message": "id is required"})
		return
	}

	// check if library exists
	lib := config.DB.Where("id = ?", id).First(&Library)
	if lib.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "library not found"})
		return
	}

	res := config.DB.Where("lib_id  = ?", id).Find(&Books)
	if res.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error finding books"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "books found", "list": Books})
}

// Search book by Title, Publisher and Author
func SearchBook(c *gin.Context) {
	var Inventory []models.BookInventory
	var data SearchBookStruct

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if data.Query == "" {
		search := config.DB.Find(&Inventory)
		if search.Error != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "could not perform search operation"})
			return
		}
	} else {
		search := config.DB.Where("title = ? OR publisher = ? OR ?=ANY(authors)", data.Query, data.Query, data.Query).Find(&Inventory)
		if search.Error != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "could not perform search operation"})
			return
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{"result": Inventory})
}

// issue book request
func IssueRequest(c *gin.Context) {
	var data IssueBookStruct
	var Inventory models.BookInventory
	var User models.Users
	var Event models.RequestEvent

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	reader, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// check if user exists
	user := config.DB.Where("id = ?", reader.ID).First(&User)
	if user.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user does not exists"})
		return
	}

	// check if book is available
	book := config.DB.Where("isbn = ?", data.ISBN).First(&Inventory)
	if book.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book does not exists"})
		return
	}

	// check if request already exists
	req := config.DB.Where("book_id = ? AND reader_id = ? AND request_type = ? AND NOT status = ?", data.ISBN, reader.ID, "issue", "rejected").First(&Event)
	if req.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "you have already requested this book"})
		return
	}

	if Inventory.AvailableCopies > 0 {
		// available
		request := config.DB.Create(&models.RequestEvent{ReaderId: reader.ID, BookId: data.ISBN, RequestDate: time.Now(), RequestType: "issue", Status: "pending"})
		if request.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error creating the request"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"message": "request has been created"})
	} else {
		// not available
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "book is not available"})
	}
}

// approve issue request
func ApproveIssueRequest(c *gin.Context) {
	var data ApproveRequestStruct
	var RequestEvent models.RequestEvent
	var Inventory models.BookInventory

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if the event exists
	event := config.DB.Where("req_id = ?", data.ReqId).First(&RequestEvent)
	if event.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "request event does not exists"})
		return
	}

	// check if user exists
	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	admin, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// approve the request
	RequestEvent.ApproverID = &admin.ID
	RequestEvent.ApprovalDate = &[]time.Time{time.Now()}[0]
	RequestEvent.Status = "approved"

	saveEvent := config.DB.Save(&RequestEvent)
	if saveEvent.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error updating the event request"})
		return
	}

	// update the availability of the book
	book := config.DB.Where("isbn = ?", RequestEvent.BookId).First(&Inventory)
	if book.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "unable to find book"})
		return
	}
	if Inventory.AvailableCopies > 0 {
		Inventory.AvailableCopies -= 1
		save := config.DB.Save(&Inventory)
		if save.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "unable to update copies"})
			return
		}

		// update the issue registry
		registry := config.DB.Create(&models.IssueRegistery{ISBN: RequestEvent.BookId, ReaderID: RequestEvent.ReaderId, IssueApproverID: admin.ID, IssueStatus: "issued", IssueDate: time.Now(), ExpectedReturnDate: time.Now().AddDate(0, 0, 7)})
		if registry.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error creating registry"})
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "request event approved"})
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "book is not available"})
	}

}

// reject request
func RejectRequest(c *gin.Context) {
	var data ApproveRequestStruct
	var RequestEvent models.RequestEvent

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// check if the event exists
	event := config.DB.Where("req_id = ?", data.ReqId).First(&RequestEvent)
	if event.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "request event does not exists"})
		return
	}

	// check if user exists
	_, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	// reject the request
	RequestEvent.Status = "rejected"

	saveEvent := config.DB.Save(&RequestEvent)
	if saveEvent.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error updating the event request"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "request event rejected"})
}

// return book request
func ReturnRequest(c *gin.Context) {
	var data IssueBookStruct
	var Inventory models.BookInventory
	var User models.Users
	var Event models.RequestEvent
	var Request models.RequestEvent

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	reader, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// check if user exists
	user := config.DB.Where("id = ?", reader.ID).First(&User)
	if user.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user does not exists"})
		return
	}

	// check if book is available
	book := config.DB.Where("isbn = ?", data.ISBN).First(&Inventory)
	if book.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book does not exists"})
		return
	}

	// check if book is issued or not
	issue := config.DB.Where("book_id = ? AND reader_id = ? AND request_type = ?", data.ISBN, reader.ID, "issue").First(&Event)
	if issue.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "you cannot return a book that has not been issued"})
		return
	}

	// check if request already exists
	req := config.DB.Where("book_id = ? AND reader_id = ? AND request_type = ? AND NOT status = ?", data.ISBN, reader.ID, "return", "rejected").First(&Request)
	if req.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "you have already requested to returned this book"})
		return
	}

	res := config.DB.Create(&models.RequestEvent{BookId: Event.BookId, ReaderId: Event.ReaderId, RequestDate: time.Now(), RequestType: "return", Status: "pending"})
	if res.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error creating the request"})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "request has been created"})
}

// approve return request
func ApproveReturnRequest(c *gin.Context) {
	var data ApproveRequestStruct
	var RequestEvent models.RequestEvent
	var IssueRegistery models.IssueRegistery
	var Inventory models.BookInventory

	err := c.ShouldBind(&data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if the event exists
	event := config.DB.Where("req_id = ?", data.ReqId).First(&RequestEvent)
	if event.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "request event does not exists"})
		return
	}

	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	admin, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// find the issue registry
	registry := config.DB.Where("isbn = ? AND reader_id = ? AND issue_approver_id = ? AND issue_status = ?", RequestEvent.BookId, RequestEvent.ReaderId, admin.ID, "issued").First(&IssueRegistery)

	if registry.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error finding the issue registry"})
		return
	}

	// update the availability of the book
	book := config.DB.Where("isbn = ?", RequestEvent.BookId).First(&Inventory)
	if book.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "unable to find book"})
		return
	}

	Inventory.AvailableCopies += 1
	save := config.DB.Save(&Inventory)
	if save.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "unable to update copies"})
		return
	}

	// approve the request
	RequestEvent.ApproverID = &admin.ID
	RequestEvent.ApprovalDate = &[]time.Time{time.Now()}[0]
	RequestEvent.Status = "approved"

	saveEvent := config.DB.Save(&RequestEvent)
	if saveEvent.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error updating the event request"})
		return
	}

	// update the issue registry
	IssueRegistery.ReturnDate = &[]time.Time{time.Now()}[0]
	IssueRegistery.ReturnApproverID = &admin.ID
	IssueRegistery.IssueStatus = "returned"

	saveRegistry := config.DB.Save(&IssueRegistery)
	if saveRegistry.Error != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error updating issue registry"})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "request event approved successfully"})

}

// Retrieve Requests
func RetrieveRequets(c *gin.Context) {
	var events []models.RequestEvent
	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	user, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// retrieve all events by user
	if user.Role == "reader" {
		// reader
		res := config.DB.Preload("BookInventory").Where("reader_id = ?", user.ID).Find(&events)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error retrieving request events"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "requests retrieved successfully", "requests": events})

	} else if user.Role == "admin" {
		// admin
		var filtered []models.RequestEvent
		res := config.DB.Preload("BookInventory").Find(&events)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error retrieving request events"})
			return
		}

		for _, value := range events {
			if value.BookInventory.LibID == user.LibID {
				filtered = append(filtered, value)
			}
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "requests retrieved successfully", "requests": filtered})

	}
}

// Retrieve Registry
func RetrieveRegistry(c *gin.Context) {
	var registry []models.IssueRegistery
	value, ok := c.Get("email")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing email in the context"})
		return
	}

	user, e := utils.FindUser(value)
	if e != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	// retrieve all events by user
	if user.Role == "reader" {
		// reader
		res := config.DB.Preload("BookInventory").Where("reader_id = ?", user.ID).Find(&registry)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error retrieving registry"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "registry retrieved successfully", "registry": registry})

	} else if user.Role == "admin" {
		// admin

		res := config.DB.Preload("BookInventory").Where("issue_approver_id = ?", user.ID).Find(&registry)
		if res.Error != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error retrieving registry"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"message": "requests retrieved successfully", "registry": registry})

	}
}
