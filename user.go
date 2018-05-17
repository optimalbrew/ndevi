//http://stackoverflow.com/questions/549/the-definitive-guide-to-form-based-website-authentication/477578#477578
//go through above and change the user struct accordingly to include throttling.

package main

import(
	"net/http"
	//"io"
	
	"errors"
	
	"golang.org/x/crypto/bcrypt"
	"appengine"
	"appengine/datastore"
	
	"encoding/gob" //for session data
	//"bytes"
)
var loginerror error = errors.New("Incorrect user name or password")
var SessSignInError error = errors.New("user not signed in")
var SessStaleError	error = errors.New("user logged in from another computer or browser. Sign in again.")

type User struct {
	Username string  //work email acts as username use lower case.
	UserFirm string	 //Employer
	Password []byte //bcrypt of password
	Token 	 []byte // bcrypt of rand string
	SessId	string //session ID to log out of old sessions
	//NumFail	 int64	//number of failed login attempts
	//FailTime time.Time // time since last failed login
	UserType string  // Ins,  RIns, or Broker	
	
}

func init(){
	gob.Register(&User{})
}

//Reads session "signin" to determine user currently logged in. Type error needs better handling. 
func SignInRecall(w http.ResponseWriter, r *http.Request)(usr User, err error){
		session, err2 := store.Get(r,"signin")
		 if err2 != nil{
        	err = SessSignInError
        	//http.Error(w, err2.Error(),500)
        	return
         }
		
		val := session.Values["curruser"]
		var sessusr = &User{}
       	
       	sessusr, ok := val.(*User)
       	if !ok{
       	err = SessSignInError      
       	//io.WriteString(w,PageTop)
        //io.WriteString(w,"<p> Invalid session signin. User not logged in.</p>")
        //io.WriteString(w,PageBot)
        return
       	} else{
       		err = nil
       	}      	
 usr = *sessusr
 
 return
}


//Users contains all users.
func UserKey (c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "Users", "default_users", 0, nil)	
}

//A method to set password, should also be used to reset password from email.
func (u *User) SetPassword(password string) {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)  
		}
	u.Password = hpass	
}


//login function. Take user name and password and try to match with hashed password.
//No response writer, we don't want to write anything before saving a session.
func Login(c appengine.Context, username string, password string) (utype string, ssid string, err2 error) {	
		q := datastore.NewQuery("Users").Filter("Username =",username).Limit(1)           
        var x []User	
        key, err := q.GetAll(c,&x) 
        if err != nil{
        	utype = ""
        	ssid = ""        		
			err2 = err
			return      		
        }
        if len(key) == 0{
        	err2 = loginerror
        	return
        }
                  		
        var y User = x[0]
        err2 = bcrypt.CompareHashAndPassword(y.Password, []byte(password))
        if err2 == nil{ //no error, so successful password match
        	utype = y.UserType  
        	ssid = RandStrings(10) //to return to signin program
        	y.SessId = ssid //
        	_, errPut := datastore.Put(c,key[0], &y) //update ssid on "Users"
        	if errPut != nil{
        		 c.Errorf("Something went wrong with the ssid put")
        	}
        }  	         	  		
	return
}

//function to check if this is the recent most value for session ID
func StaleLog(c appengine.Context, user User)(error){
	qRcnLog := datastore.NewQuery("Users").Filter("Username =",user.Username).Limit(1)
	
	//first grab the key only										   
	var UserCred []User
	Keys,errSessID := qRcnLog.KeysOnly().GetAll(c,&UserCred)
	if errSessID != nil{
		c.Errorf("couldn't get user's key from store")
		return SessStaleError
	}
	//then instead of query, do an actual get. This does not require index.
	var y User
	errSessId2 := datastore.Get(c,Keys[0], &y)
	if errSessId2 != nil{
		c.Errorf("couldn't get latest ssid from store")
	}
	
	//do the comparison
	//c.Errorf(y.SessId) //from store for testing
	//c.Errorf(user.SessId) //from cookie for testing
	if y.SessId != user.SessId{	
		c.Errorf("user sess ID does not match login creds")
		return SessStaleError
	}else{
	return nil
	}
}

