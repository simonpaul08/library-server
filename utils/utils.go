package utils

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"os"
	"project/libraryManagement/config"
	"project/libraryManagement/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/skip2/go-qrcode"
	gomail "gopkg.in/mail.v2"
)

// find library
func FindLibrary(name string) (*models.Library, error) {
	var library models.Library
	res := config.DB.First(&library, "name = ?", name)

	if res.Error != nil {
		return nil, res.Error
	}

	return &library, nil
}

// find user
func FindUser(email any) (*models.Users, error) {
	var user models.Users
	res := config.DB.Preload("Library").First(&user, "email = ?", email)

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil

}

// random number
func getRandNum() (string, error) {
	nBig, e := rand.Int(rand.Reader, big.NewInt(8999))
	if e != nil {
		return "", e
	}
	return strconv.FormatInt(nBig.Int64()+1000, 10), nil
}

// send email
func SendMail(email string, message string, subject string) error {


	from := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")

	m := gomail.NewMessage()

	m.SetHeader("From", from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// setting smtp server
	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // for ssl certificate, should be marked false in production

	// sending email
	sendEmail := d.DialAndSend(m)
	if sendEmail != nil {
		fmt.Println(sendEmail)
		return errors.New("couldn't send mail")
	}

	return nil
}

// send otp for verification
func SendOTP(email string) error {
	str, err := getRandNum()
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("error generating otp")
	}

	emailText := fmt.Sprintf("Hey, Please use this otp to log in to your account: %s", str)

	result := SendMail(email, emailText, "OTP for verification")
	if result != nil {
		return errors.New("error sending email")
	}

	// update the otp in the database table
	user, err := FindUser(email)
	if err != nil {
		return errors.New("error finding user")
	}

	user.OTP = str
	save := config.DB.Save(&user)
	if save.Error != nil {
		return errors.New("error")
	}

	return nil
}

// verify otp and send user back
func VerifyOTP(email, otp string) (*models.Users, error) {
	user, err := FindUser(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// compare otp
	if otp != user.OTP {
		return nil, errors.New("invalid otp")
	}

	user.OTP = ""
	result := config.DB.Save(&user)
	if result.Error != nil {
		return nil, errors.New("error updating the user")
	}

	return user, nil
}

// generate token
func GenerateToken(user *models.Users) (string, error) {
	secret := []byte(os.Getenv("SECRET"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  user.Email,
		"role":   user.Role,
		"expiry": time.Now().Add(time.Second * time.Duration(5000)).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generate qr code 
func GenerateQR(text string) ([]byte, error){
	var png []byte
	png, err := qrcode.Encode(text, qrcode.Medium, 256)

	if err != nil {
		return nil, err
	}

	return png, err
}
