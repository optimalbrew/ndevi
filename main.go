//This is the main file. With all the handler files and functions defined separately.
//dev_appserver.py --clear_datastore=yes proto2
//dev_appserver.py --smtp_host=smtp.mail.yahoo.com --smtp_port=587 --smtp_user=deleted --smtp_password=deleted proto2
package main 

import(
		"net/http"
	//	"io"
		"log"
	//	"strconv"
	//	"html/template"
	//	"fmt"
    // 	"time"
    //	"errors"
    //    "math"
    //    "sort"
    //	"bytes"
    			
    //   "appengine"
    //   "appengine/datastore"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "github.com/gorilla/securecookie"   
)
//generate random keys for authentication and encryption
var sessAuth = securecookie.GenerateRandomKey(64)
var sessEncr = securecookie.GenerateRandomKey(32)

var store = sessions.NewCookieStore(sessAuth,sessEncr)

var r = mux.NewRouter()        
        
//a function type, passed on to logPanics
type HandFn func(http.ResponseWriter, *http.Request)


//type function of type HandFn as input, get the same as output wrapped
//inside a panic,defer,recover.
func logPanics(somefunc HandFn) HandFn{
	return func(writer http.ResponseWriter, request *http.Request){
		defer func(){
			if x:=recover(); x!= nil{
				log.Printf("[%v] caught panic: %v", request.RemoteAddr, x)
			}
		}()
		
		somefunc(writer, request)
	}
}



//The main function that'll handle all request.
func init(){
	
	r.HandleFunc("/", logPanics(root))    // "/"
	r.HandleFunc("/request/{key}", logPanics(CreateRFQ))      // "/read RFQ form input"
	r.HandleFunc("/regreq", logPanics(RegReq))      // "/registration"
	r.HandleFunc("/signup", logPanics(SignUp))      // "/signup"
	r.HandleFunc("/reslink", logPanics(ResetChangePage)) //link to reset or change passwd
	r.HandleFunc("/resetreq", logPanics(ResetReq))      // "/request reset"
	r.HandleFunc("/resetpwd", logPanics(ResetPassword))      // "/reset password"
	r.HandleFunc("/signin", logPanics(SignIn))      // "/SignIn"	
	r.HandleFunc("/signout", logPanics(Logout))      // "/sign out"
	r.HandleFunc("/inspage", logPanics(InsPage))	//inspage show form for RFQ	
	r.HandleFunc("/treaty", logPanics(TreatyPage))	//inspage show form for RFQ	Trety
	r.HandleFunc("/handleroot", logPanics(handleRoot))      // "/root handle for file upload"
	r.HandleFunc("/upload", logPanics(handleUpload))      // "/ handle the upload of files
	r.HandleFunc("/uploadrins", logPanics(handleUploadRins))      // "/ handle the upload of files
	r.HandleFunc("/review", logPanics(ReviewRFQ))      // "/ review RFQ after adding files"
	r.HandleFunc("/reviewrins", logPanics(RevRinsFile))      // "/ view summary after file upload
	r.HandleFunc("/serve/{key}", logPanics(handleServe))      // "/"serve file from blobstore
	r.HandleFunc("/serverins/{rinsId}/{fileId}", logPanics(handleServeRins))      // "/"serve file from blobstore
	r.HandleFunc("/deleteFileIns/{key}", logPanics(RemoveFileIns))      // "/"serve file from blobstore
	r.HandleFunc("/deleteFileRins/{key}", logPanics(RemoveFileRins))      // "/"serve file from blobstore	
	r.HandleFunc("/addrins", logPanics(AddRins))      // "/ remove or add more Rinsurers from RFQ "
	r.HandleFunc("/sharerfq", logPanics(ShareRFQ))      // "/ save and send email "
	r.HandleFunc("/insdboard", logPanics(InsurerDashboard))      // "/ insurer dashboard "
	r.HandleFunc("/search", logPanics(SearchBase))      // "/ insurer dashboard "
	r.HandleFunc("/rinsdboard", logPanics(RIDashboard))      // "/ Reinsurer dashboard "
	r.HandleFunc("/inslocrfq/{key}", logPanics(InsLocRFQ))      // "/ locate and display RFQ (Ins)"	
	r.HandleFunc("/rinslocrfq/{key}", logPanics(RInsLocRFQ))      // "/ locate and display RFQ (Rins)"
	//r.HandleFunc("/modifyrfq", logPanics(ModifyRFQ))      // "/ Show revision form for current RFQ "
	r.HandleFunc("/modalert", logPanics(ModAlert))      // "/ Show revision form for current RFQ "
	r.HandleFunc("/reviserfq", logPanics(ReviseRFQ))      // "/ delete located RFQ "		
	r.HandleFunc("/deleterfq", logPanics(MarkRFQdel))      // "/ mark an RFQ as deleted "	
	r.HandleFunc("/recoverrfq", logPanics(RecoverRFQdel))      // "/ recover RFQ marked as deleted "	
	//r.HandleFunc("/batchdeleterfq", logPanics(DeleteBatchRFQ))      // "/ batch delete RFQs "
	r.HandleFunc("/rinsinit", logPanics(RinsInitResp))      // "/ Rins' initial form read "	
	r.HandleFunc("/rinsrespmsg", logPanics(RinsRespMsg))      // "/ Rins' actual response sent "
	r.HandleFunc("/pm/{key}", logPanics(pm))      // "/ Ins' pm to Rins "	
	http.Handle("/",r)	
}



/*TODO
6: Error handling everywhere and logs.
8. Run security scan on appengine.
*/
