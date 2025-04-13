package models

import (
	"time"
	pq "github.com/lib/pq"
)

type Library struct {
	ID   uint `json:"id" gorm:"primaryKey"`
	Name string	`json:"name" gorm:"unique"`
}

type Users struct {
	ID            uint  	`json:"id" gorm:"primaryKey"`
	Name          string  	`json:"name"`
	Email         string  	`json:"email" gorm:"unique"`
	ContactNumber string    `json:"contactNumber"`
	Role          string  	`json:"role"`
	LibID         uint  	`json:"libId"`
	Library       Library 	`gorm:"foreignKey:ID;references:LibID"`
	OTP			  string 	`json:"otp"`
}

type BookInventory struct {
	ISBN            uint         	`json:"isbn" gorm:"primaryKey"`
	Title           string         	`json:"title"`
	Authors         pq.StringArray 	`json:"authors" gorm:"type: varchar(200)[]"`
	Publisher       string         	`json:"publisher"`
	Version         string         	`json:"version"`
	TotalCopies     uint           	`json:"totalCopies"`
	AvailableCopies uint           	`json:"availableCopies"`
	QrCode 			[]byte			`json:"qrCode"`
	LibID           uint         	`json:"libID"`
	Library         Library        	`gorm:"foreignKey:ID;references:LibID"`
}		

type RequestEvent struct {
	ReqId         uint        	`json:"reqId" gorm:"primaryKey"`
	BookId        uint        	`json:"bookId"`
	ReaderId      uint        	`json:"readerId"`
	RequestDate   time.Time     `json:"requestDate"`
	ApprovalDate  *time.Time    `json:"approvalDate"`
	ApproverID    *uint        	`json:"approverId"`
	RequestType   string        `json:"requestType"`
	Status		  string		`json:"status"`
	BookInventory BookInventory `gorm:"foreignKey:ISBN;references:BookId"`
	Users         Users         `gorm:"foreignKey:ID;references:ReaderId,ApproverID"`
}

type IssueRegistery struct {
	IssueID         	uint    		`json:"issueId" gorm:"primaryKey"`
	ISBN            	uint    		`json:"isbn"`
	ReaderID        	uint    		`json:"readerId"`
	IssueApproverID 	uint    		`json:"issueApproverId"`
	IssueStatus     	string    		`json:"issueStatus"`
	IssueDate       	time.Time 		`json:"issueDate"`
	ExpectedReturnDate	time.Time		`json:"expectedReturnDate"`
	ReturnDate			*time.Time		`json:"returnDate"`
	ReturnApproverID	*uint			`json:"returnApproverId"`
	BookInventory 		BookInventory	`gorm:"foreignKey:ISBN;references:ISBN"`
	Users				Users			`gorm:"foreignKey:ID;references:ReaderID,IssueApproverID,ReturnApproverID"`
}
