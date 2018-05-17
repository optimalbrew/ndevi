//testing page for user management. Signup, reset password, and login.
//Still needs to be integrated with email confirmation (with one time code) to confirm.
//dev_appserver.py --smtp_host=smtp.mail.yahoo.com --smtp_port=587 --smtp_user= //deleted --smtp_password= deleted  proto


package main

import(
	"net/http"
	"strings"
	//"strconv"
	"fmt"
	"io"
	//"bytes"
	//"time"
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/sessions"
//	"github.com/gorilla/securecookie"
//	"encoding/gob"
)
//should check for these errors
//type M map[string]interface{}

//func init(){
//	gob.Register(&M{})
//}
	

//Registration request using email. Only take email as input and sends email/token. 
func RegReq(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprint(w, PageTop) //Alternative to io.WriteString
	//Now to save the data
	c := appengine.NewContext(r)
	
		username := strings.ToLower(r.Form.Get("exampleInputEmail1"))		
		//Generate random token
		randStr := RandStrings(20)
		hashTok,_ := bcrypt.GenerateFromPassword([]byte(randStr), bcrypt.DefaultCost)
		user := User{
					Username: username,
					Token: hashTok, //will serve to confirm email.
				}		
		c.Errorf("for testing only print randStr [in RegReq]")
		c.Errorf(randStr)
		 
		q := datastore.NewQuery("Users").Filter("Username =",user.Username)                   
        var x []User	
        key, errX := q.GetAll(c,&x)
        if errX != nil{
        	c.Errorf("could not GetAll in RegReq")
        	return
        }
        numEntries := len(key)
        if numEntries > 1{
        	c.Errorf("multiple registration for same username")
        	//Send email
        		
               	s1 := []string{user.Username, "multiple"}     
                rcpt := []string{"support@ndevi.com"}  //email rcpt list
				//email struct
                msg := &mail.Message{
                Sender: //deleted
                To: rcpt,
                Subject: "ndevi Registration multiplicity",
                Body: strings.Join(s1, ""),                
                }
                
                if errMail := mail.Send(c,msg); errMail != nil {
                	c.Errorf("send mail error")
                }
               	io.WriteString(w, register)
                io.WriteString(w, PageBot)	
        	return
        }
        
        if numEntries == 0{  //user not found, permit addition and send email.
        		
        		keyUser := datastore.NewIncompleteKey(c, "Users", UserKey(c))
        		_, err := datastore.Put(c, keyUser, &user)
        		if err != nil {
                	http.Error(w, err.Error(), http.StatusInternalServerError)
                	return
                } 
                
                //Send email
               	s1 := []string{msgbody1, randStr}     
                rcpt := []string{username}  //email rcpt list
				//email struct
                msg := &mail.Message{
                Sender: //deleted
                To: rcpt,
                Subject: "ndevi Registration confirmation",
                Body: strings.Join(s1, ""),                
                }
                
                if errMail := mail.Send(c,msg); errMail != nil {
                	io.WriteString(w,"<p>Oops!Please try again or report problem.</p>")
                }                
                
                io.WriteString(w, register)  //register is a constant in templates.go                          	                
                io.WriteString(w, PageBot)										
        		return
        } else{   //user previously registered. Check if registration was complete.            	
        		
        		userFound := x[0]
        		keyFound := key[0]
        		
        		if userFound.UserType != ""{//previous complete registration
        			//Don't do anything
        			//But for now, display the same form. 
 	                io.WriteString(w, register)                                                   	
    	    		io.WriteString(w, PageBot)
        			return
        		
        		}else{ //incomplete registration, send email again.
        			_, errput := datastore.Put(c, keyFound, &user)
        			if errput != nil{
        				http.Error(w, errput.Error(), http.StatusInternalServerError)
        				return
        			}
        			//Send email
					s1 := []string{msgbody1, randStr}     
					rcpt := []string{username}  //email rcpt list
					//email struct
					msg := &mail.Message{
					Sender: //deleted
					To: rcpt,
					Subject: "ndevi Registration confirmation",
					Body: strings.Join(s1, ""),                
					}
				
					if errMail := mail.Send(c,msg); errMail != nil {
						io.WriteString(w,"<p>Oops!Please try again or report problem.</p>")
					}                
				
					io.WriteString(w, register)  //register is a constant in templates.go                          	                
					io.WriteString(w, PageBot)										
					return
        				
        		
        		}
        }	
        	
		
}

const msgbody1 = `Thank you for registering with ndevi. To complete your registration please copy and paste the token below in the sign up form where you can set your own password. If you have closed the browser tab, the signup page is available at https://www.ndevi.com/signup

To reduce the chances of phishing attacks, please check that the above link actually points to what it says, or use your browser's history.

Your single use token is:

`



func SignUp(w http.ResponseWriter, r *http.Request) {
       
       	r.ParseForm()
       
			//Now to save the data
		c := appengine.NewContext(r)

		username := strings.ToLower(r.Form.Get("exampleInputEmail1"))
		comp := strings.Split(username,"@")
		token := r.Form.Get("token")
		password1 := r.Form.Get("exampleInputPassword1")  //from form as string, SetPassword will read as []btye
		password2 := r.Form.Get("exampleInputPassword2")		
		usertype := r.Form.Get("RadioProd")  //needs to be converted from string (form) to int64
		
		if password1 != password2{				
			io.WriteString(w, "<p>Passwords do not match. Please try again.</p>")
			io.WriteString(w, register)
			io.WriteString(w, PageBot)
			return		
		}
				
		user := User{
				Username: username,
				UserFirm: comp[1],
				UserType: usertype, 
				}		
		 
		 //now to find the user and match the token
		 q := datastore.NewQuery("Users").Filter("Username =",user.Username).Limit(1)  
         
         for t := q.Run(c); ; {
        	var x User	
        	key, err := t.Next(&x) //go to next match
        	
        	if err != nil {  //user not found,
        			//no token to match, no user. 
        			fmt.Fprint(w, PageTop) 
        			io.WriteString(w,"<p>An error occurred. Please register again and get a new token.</p>")
        			io.WriteString(w, emailReg)
        			io.WriteString(w, PageBot)
        			break
        	} else{
        	 
        		//compare token with hash
        		err2 := bcrypt.CompareHashAndPassword(x.Token, []byte(token)) 	
        		if err2 !=nil {
        			fmt.Fprint(w, PageTop) 
        			io.WriteString(w,"<p>Invalid token. Please enter the token again.</p>")
        			io.WriteString(w, register)
        			io.WriteString(w, PageBot)
        			break
        		}
        		
        		user.SetPassword(password1) //generate bcrypt password hash
        		//to replace one time token with something else
        		s :=[]byte(RandStrings(8)) 
        		user.Token,_ = bcrypt.GenerateFromPassword(s, bcrypt.DefaultCost)
        		//keyUser := datastore.NewIncompleteKey(c, "Users", UserKey(c))
        		_, err := datastore.Put(c, key, &user)
        		if err != nil {
                	http.Error(w, err.Error(), http.StatusInternalServerError)
                }
                fmt.Fprint(w, PageTop)                 
                io.WriteString(w,"<p>Registration for <b>"+ user.Username + " </b> complete. Please sign in.</p>")
                
                io.WriteString(w, retry)       
                io.WriteString(w, PageBot)										
        		break
        	}         		
        }	
		
}

//Log using in, start a session and then write anything to response writer, not before.
func SignIn(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        c := appengine.NewContext(r)
  		username := strings.ToLower(r.Form.Get("exampleInputEmail1"))
		password := r.Form.Get("exampleInputPassword1")  //from form as string
        
       	user := User{
       			Username: username,
       		} 
		//Login function. Returns usertype, ssid, and error.
		userType, ssid, err := Login(c, user.Username, password)
        if err != nil{        		
        	io.WriteString(w, PageTopCent) 
        	io.WriteString(w, "<p>Incorrect username or password.</p>")              
        	//io.WriteString(w, retry)
        	io.WriteString(w, PageBotCent) 
        	return	    	        	
        } 
		user.UserType = userType //Rins or Ins or Broker
		user.SessId = ssid
		
		//start a session, save it, and then call any functions that write stuff
		session, err2 := store.Get(r,"signin")
		if err2!=nil{
			http.Error(w,err2.Error(),500)
			return
		}
	
		session.Options  = &sessions.Options{
			Path: "/",
			MaxAge: 0, //not persistent, session cookie //use 86400*0.5 for 12 hours
			HttpOnly: true,
		}       	
		//set session value        
		session.Values["curruser"] = user
		
		//clear any previous session cookies
		//clear cookie for last remembered RFQ info        	
		session2, err2 := store.Get(r,"sessIns")
		if err2 != nil{
			http.Error(w,err2.Error(),500)
			return
		}
		session2.Options  = &sessions.Options{
			Path: "/",
			MaxAge: -1, //delete the cookie
			HttpOnly: true,
		}    
		//clear cookie for last remembered RFQ Location info
		session3, err3 := store.Get(r,"sessLocKey")
		if err3 != nil{
			http.Error(w,err3.Error(),500)
			return
		}
		session3.Options  = &sessions.Options{
			Path: "/",
			MaxAge: -1, //delete the cookie
			HttpOnly: true,
		}    
		session4, err4 := store.Get(r,"sessRins")
		if err4 != nil{
			http.Error(w,err4.Error(),500)
			return
		}
		session4.Options  = &sessions.Options{
			Path: "/",
			MaxAge: -1, //delete the cookie
			HttpOnly: true,
		}   
		//save new signin cookie remove the others               
		sessions.Save(r,w)  //note, not (w,r)}
		//redirect to appropriate dashboard
		if userType != "RIns"{
			//Not reinsurer => insurer or reinsurance broker.
			http.Redirect(w,r,"/insdboard", http.StatusFound)	
						
		}else{//reinsurer
		http.Redirect(w,r,"/rinsdboard", http.StatusFound)
		}
}



//Password reset, send email token and set new token.
func ResetReq(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprint(w, PageTop) //Alternative to io.WriteString
	//Now to save the data
	c := appengine.NewContext(r)
	
		username := strings.ToLower(r.Form.Get("exampleInputEmail1"))
		
		
		//Generate random token
		randStr := RandStrings(20)
		hashTok,_ := bcrypt.GenerateFromPassword([]byte(randStr), bcrypt.DefaultCost)
		user := User{
					Username: username,
					//Token: hashTok, //will serve to confirm email.
				}		
		 
		 q := datastore.NewQuery("Users").Filter("Username =",user.Username).Limit(1)  
         
         for t := q.Run(c); ; {
        	var x User	
        	key, err := t.Next(&x) //go to next match
        	
        	if err != nil {  //user not found, don't do anything. Display same form.
        		
                io.WriteString(w, emailReset)                            	                
                io.WriteString(w, PageBot)										
        		break
        	} else{ 
        		x.Token = hashTok
        		_, err := datastore.Put(c, key, &x)
        		if err != nil {
                	http.Error(w, err.Error(), http.StatusInternalServerError)
                } 
        		     	
        		s2 := []string{msgbody2,randStr}
               
                rcpt := []string{username}  //email rcpt list
                msg := &mail.Message{
                Sender: //deleted
                To: rcpt,
                Subject: "ndevi Password Reset",
                Body: strings.Join(s2, ""),                
                }
                
                if errMail := mail.Send(c,msg); errMail != nil {
                	io.WriteString(w,"<p>Oops! We have a problem. Please email support and we will sort it out.</p>")
                }                
                
                io.WriteString(w, emailReset)                            	
                
        	
        		io.WriteString(w, PageBot)
        	break
        	}	
        }	
		
}


const msgbody2 = `Dear User,
We received a request for a password reset for the user with this email ID on www.ndevi.com. If you did not request this change, then please disregard this email. If you wish to change your password then please enter the following single use token in the password reset form. If you have closed the browser tab, the reset page is available at https://www.ndevi.com/resetpwd

To reduce the chances of phishing attacks, please  check that the above link actually points to what it says, or use your browser's history.

Your single use token is:
 
`

func ResetChangePage(w http.ResponseWriter, r *http.Request) {
io.WriteString(w,PageTop)
io.WriteString(w,emailPwdResetForm)
io.WriteString(w,PageBot)
}



func ResetPassword(w http.ResponseWriter, r *http.Request) {
       
       	r.ParseForm()
       
		fmt.Fprint(w, PageTop) //Alternative to io.WriteString
			//Now to save the data
		c := appengine.NewContext(r)

		username := strings.ToLower(r.Form.Get("exampleInputEmail1"))
		//comp := strings.Split(username,"@")
		token := r.Form.Get("token")
		password1 := r.Form.Get("exampleInputPassword1")  //from form as string, SetPassword will read as []btye
		password2 := r.Form.Get("exampleInputPassword2")		
		//usertype := r.Form.Get("RadioProd")  //needs to be converted from string (form) to int64
		
		if password1 != password2{				
			io.WriteString(w, "<p>Passwords do not match. Please try again.</p>")
			io.WriteString(w, register)
			io.WriteString(w, PageBot)
			return		
		}
				
		user := User{
				Username: username,
				//UserFirm: comp[1],
				//UserType: usertype, 
				}		
		 
		 //now to find the user and match the token
		 q := datastore.NewQuery("Users").Filter("Username =",user.Username).Limit(1)  
         
         for t := q.Run(c); ; {
        	var x User	
        	key, err := t.Next(&x) //go to next match
        	
        	if err != nil {  //user not found,
        			//no token to match, no user. 
        			io.WriteString(w,"<p>An error occurred. Please request a new token.</p>")
        			io.WriteString(w, emailPwdResetForm)
        			io.WriteString(w, PageBot)
        			break
        	} else{
        	 
        		//compare token with hash
        		err2 := bcrypt.CompareHashAndPassword(x.Token, []byte(token)) 	
        		if err2 !=nil {
        			io.WriteString(w,"<p>Invalid token. Please enter the token again.</p>")
        			io.WriteString(w, emailPwdResetForm)
        			io.WriteString(w, PageBot)
        			break
        		}
        		
        		x.SetPassword(password1) //generate bcrypt password hash
        		//to replace one time token with something else
        		s :=[]byte(RandStrings(20)) 
        		x.Token,_ = bcrypt.GenerateFromPassword(s, bcrypt.DefaultCost)
        		//keyUser := datastore.NewIncompleteKey(c, "Users", UserKey(c))
        		_, err := datastore.Put(c, key, &x)
        		if err != nil {
                	http.Error(w, err.Error(), http.StatusInternalServerError)
                }
                                
                io.WriteString(w,"<p>Password changed for <b>"+ x.Username + " . Please sign in.</p>")
                
                io.WriteString(w, retry)       
                io.WriteString(w, PageBot)										
        		break
        	}         		
        }			
}

func Logout(w http.ResponseWriter, r *http.Request){
	//remove closed RFQs
	c := appengine.NewContext(r)
	errBatch := DeleteBatchRFQ(w,r)
	if errBatch != nil{
			c.Errorf("problem in batch deletion in signout")			
	}
	session1, err := store.Get(r,"signin")
		 if err != nil{
        	http.Error(w,err.Error(),500)
        	return
         }
	session1.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: -1, //delete the cookie
        	HttpOnly: true,
        }    
        
    session2, err2 := store.Get(r,"sessIns")
		 if err2 != nil{
        	http.Error(w,err2.Error(),500)
        	return
         }
	session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: -1, //delete the cookie
        	HttpOnly: true,
        }    

	session3, err3 := store.Get(r,"sessLocKey")
	if err3 != nil{
        	http.Error(w,err3.Error(),500)
        	return
         }
	session3.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: -1, //delete the cookie
        	HttpOnly: true,
        }    
        
    session4, err4 := store.Get(r,"sessRins")
	if err4 != nil{
        	http.Error(w,err4.Error(),500)
        	return
         }
	session4.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: -1, //delete the cookie
        	HttpOnly: true,
        }
        
    session5, err5 := store.Get(r,"sessResp")
	if err5 != nil{
        	http.Error(w,err4.Error(),500)
        	return
         }
	session5.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: -1, //delete the cookie
        	HttpOnly: true,
        }
	 //save the session               
    sessions.Save(r,w)  //note, not (w,r)}
	//show sign in page again
	io.WriteString(w, PageTopCent) 
    io.WriteString(w, "<p>User logged out or signed in from another computer/browser.</p>")              
        	//io.WriteString(w, retry)
    io.WriteString(w, PageBotCent) 
}


