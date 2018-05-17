//reinsurers interface functions

package main

import(
	"net/http"
//	"fmt"
	"io"
	
	"strings"
	//"strconv"
	"html/template"
	"appengine"
	"appengine/datastore"
	//"appengine/blobstore"

	"time"	
	"github.com/gorilla/sessions"
	"github.com/gorilla/mux"
) 

//function to locate and respond to RFQs
//WARNING: Will pull info from RFQ, but not store a lot of it.
// Only things of relevance (including some for querying) to Rins will be stored
//So need to distinguish between what's stored in the datastore versus what's stored in
// temporary session cookies.
func RInsLocRFQ(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	usr, err := SignInRecall(w,r)	
	if err != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		c.Errorf("error reading session values")
		return
	}
	
	vars := mux.Vars(r)
	rfqId := vars["key"]
	
	//locate the RFQ and to ensure it matches both the rfq ID and the user's ID
	var rinsFindrfq []RInsRFQList
	q := datastore.NewQuery("RInsList").Filter("RIUsername =",usr.Username).Filter("RFQId =", rfqId)           
        key1, err2 := q.GetAll(c, &rinsFindrfq)
        if err2 != nil{
        	c.Errorf("RFQ not found")
        	return
        }
        if len(key1) >1{
        	c.Errorf("multiple entries in RinsList in RInsLocRFQ")
        	return
        }	
   	//the located underlying key for the RFQ itself,
    keyLoc := rinsFindrfq[0].ReqRcvd
    rfq := new(RFQ) //not make
    err3 := datastore.Get(c, keyLoc, rfq)
    if err3 != nil{
    	c.Errorf("could not Get RFQ in RInsLocRFQ")
    	return
    }
    
    //copy pertinent information from both structs (RFQ and RinsRFQList) to a temporary RinsRFQ for
    //use in template to display information and to save in session.
    RinsRFQ := RInsRFQList{
    			//where ever possible use from RinsList itself
    			RFQId: rinsFindrfq[0].RFQId,	//	string //store
				RIUsername: rinsFindrfq[0].RIUsername,	 //this is the primary Rins contact
				//RIUsername is out of order from RFQ struct 
				//NO! IntID		string 
				//NO! Iterator	int	
				Insurer: rfq.Insurer,	//string //store, it's easy.
				InsEmpID: rfq.InsEmpID,	//string //store
				//The following from Insurers RFQ
				ClientName: rfq.ClientName,	//string `datastore:",noindex"`
				InsCat: rfq.InsCat,	//	string //store
				NewCust: rfq.NewCust,	//	string `datastore:",noindex"`	
				InsType: rfq.InsType,	//	string `datastore:",noindex"`
				InsVal: rfq.InsVal, 	//string `datastore:",noindex"`
				InsDed:	rfq.InsDed,	//string `datastore:",noindex"`
				Method: rfq.Method,
				EstPrem: rfq.EstPrem,
				Period:	rfq.Period,
				Brokerage: rfq.Brokerage, //just for brokers
				InsRemarks: rfq.InsRemarks, //string	  `datastore:",noindex"`
				FileNames: rfq.FileNames, //	[]string  `datastore:",noindex"`
				FileBlobKeys: rfq.FileBlobKeys, //[]appengine.BlobKey `datastore:",noindex"`//[]BlobKeys for all attached files.
				//NO! ReInsList	[]string  
				//Not now RFQDate: 		time.Time
				RFQStrTime: rfq.RFQStrTime,	//string `datastore:",noindex"`
				//NO! Status		int64 //1:saved, 2: submitted/in process, 3: accepted
				//NO! Modified	string //"new", "modified", "modification emailed"		
				//Not now SecondRins string	//this is the secondary who can also make modifications. Later.
				//Not now Observers	[]string //list of User's with observer status. Later.	
				ReqRcvd: rinsFindrfq[0].ReqRcvd,	//*datastore.Key  `datastore:",noindex"`//keys of RFQs with response required	
				RespStatus: rinsFindrfq[0].RespStatus, //int64 //1: response needed, 0 done.
				RespDate: rinsFindrfq[0].RespDate, //have responded at least once.
				RinsFileNames: rinsFindrfq[0].RinsFileNames, //	[]string `datastore:",noindex"` //files attached by Rins
				RinsFileBlobs: rinsFindrfq[0].RinsFileBlobs,	//[]appengine.BlobKey `datastore:",noindex"` //Rins' file blobkeys
				Limit: rinsFindrfq[0].Limit,	//	[]string `datastore:",noindex"`//Capacity offerred []s allow change history
				Premium: rinsFindrfq[0].Premium,	//	[]string `datastore:",noindex"`//[]s allow change history
				Commission: rinsFindrfq[0].Commission,	//[]string `datastore:",noindex"`//[]s allow change history
				Conversation: rinsFindrfq[0].Conversation, //[]string `datastore:",noindex"`//conversation. Combined notes so they show up as dialogue.
				ConvAuth: rinsFindrfq[0].ConvAuth,
				Private: rinsFindrfq[0].Private,		//string	`datastore:",noindex"`//notes from Rins to self
				}
    
    
    //store the temporary struct in a session cookie
    session, err := store.Get(r,"sessRins")
        if err!=nil{
        	http.Error(w,err.Error(),500)
        	return
        }       
    session.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session.Values["CurrRinsRFQ"] = RinsRFQ
    
    //store the located key from RinsListRFQ in a session cookie as well, will be
    //useful for storing changes on the response POST
    session1, errKey := store.Get(r,"sessLocKey")
	if errKey != nil{
        	http.Error(w,errKey.Error(),500)
        	return
    }	
    session1.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }
    cookieForLocRinsRFQkey := sessLocKey{
						RFQId: RinsRFQ.RFQId,
						RfqKey: key1[0],
					}             
    session1.Values["Key"] = cookieForLocRinsRFQkey        
    sessions.Save(r,w)  //note, not (w,r)
    
  	t := template.New("RinsMainRFQ.html")
	t  = template.Must(t.ParseFiles("templates/RinsMainRFQ.html"))
	err = t.Execute(w,RinsRFQ)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}
}


//RIns Dashboard
func RIDashboard(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)	
	usr, err0 := SignInRecall(w,r)	
	if err0 != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		c.Errorf("error reading session values")
		return
	}	
	//check to verify this is the most recent login	
	errStale := StaleLog(c,usr)
	if errStale != nil{
		http.Redirect(w,r,"/signout",http.StatusFound)
		c.Errorf("stale login") //remove this later. 
		return
	}	
	if usr.UserType != "RIns"{ //Ins and broker cannot view rinsdboard
		c.Errorf("Only Rins can access Rins dboard")
		io.WriteString(w, PageTop)
		io.WriteString(w, "Invalid user type.")
		io.WriteString(w, PageBot)
		return
	}
	
	//Find the associated RFQ's	 for Status 1
		var RIrfqListStat1 []RInsRFQList //these are the relevant RFQs
	  	q1 := datastore.NewQuery("RInsList").Filter("RIUsername =",usr.Username).
	  						Filter("RespStatus =", 1).Order("-RFQDate")//.Offset(10).Limit(10)           
        //GetAll is a method for type query, and needs &dst, not dst. Unlike Get or GetMulti
        _, err1 := q1.GetAll(c, &RIrfqListStat1)
        if err1 != nil{
        	c.Errorf("Query Error")
        }	
        
		//The only thing different here would be the Key of the insurer's RFQ
		//Find the relevant RFQ's.
		numRFQs1 := len(RIrfqListStat1)
		relevantKeys1 := make([]*datastore.Key, numRFQs1)
		for i :=0;i<numRFQs1;i++{
			relevantKeys1[i] = RIrfqListStat1[i].ReqRcvd
		}
				
		relevantRFQs1 := make([]RFQ,numRFQs1) 
		//Get all works with the just dst, not &dst (or perhaps make returns a pointer, so one is not rqd.)
		errGet1 := datastore.GetMulti(c, relevantKeys1, relevantRFQs1)
		if errGet1 != nil{
			c.Errorf("Relevant RFQs not retrieved")
		}
		//the iterator for table rows
		for i:=0; i<len(relevantRFQs1); i++{
			relevantRFQs1[i].Iterator = i+1
		}
		
	//Find the associated RFQ's	for Status 2
		var RIrfqListStat2 []RInsRFQList //these are the relevant RFQs
	  	q2 := datastore.NewQuery("RInsList").Filter("RIUsername =",usr.Username).
	  						Filter("RespStatus =", 2).Order("-RFQDate")//.Offset(10).Limit(10)           
        //GetAll is a method for type query, and needs &dst, not dst. Unlike Get or GetMulti
        _, err2 := q2.GetAll(c, &RIrfqListStat2)
        if err2 != nil{
        	c.Errorf("Query Error")
        }	
        
		//The only thing different here would be the Key of the insurer's RFQ
		//Find the relevant RFQ's.
		numRFQs2 := len(RIrfqListStat2)
		relevantKeys2 := make([]*datastore.Key, numRFQs2)
		for i :=0;i<numRFQs2;i++{
			relevantKeys2[i] = RIrfqListStat2[i].ReqRcvd
		}
				
		relevantRFQs2 := make([]RFQ,numRFQs2) 
		//Get all works with the just dst, not &dst (or perhaps make returns a pointer, so one is not rqd.)
		errGet2 := datastore.GetMulti(c, relevantKeys2, relevantRFQs2)
		if errGet2 != nil{
			c.Errorf("Relevant RFQs not retrieved")
		}
		//the iterator for table rows
		for i:=0; i<len(relevantRFQs2); i++{
			relevantRFQs2[i].Iterator = i+1
		}	
	//Find the associated RFQ's	 for Status 3
		var RIrfqListStat3 []RInsRFQList //these are the relevant RFQs
	  	q3 := datastore.NewQuery("RInsList").Filter("RIUsername =",usr.Username).
	  						Filter("RespStatus =", 3).Order("-RFQDate")//.Offset(10).Limit(10)           
        //GetAll is a method for type query, and needs &dst, not dst. Unlike Get or GetMulti
        _, err3 := q3.GetAll(c, &RIrfqListStat3)
        if err3 != nil{
        	c.Errorf("Query Error")
        }	
        
		//The only thing different here would be the Key of the insurer's RFQ
		//Find the relevant RFQ's.
		numRFQs3 := len(RIrfqListStat3)
		relevantKeys3 := make([]*datastore.Key, numRFQs3)
		for i :=0;i<numRFQs3;i++{
			relevantKeys3[i] = RIrfqListStat3[i].ReqRcvd
		}
				
		relevantRFQs3 := make([]RFQ,numRFQs3) 
		//Get all works with the just dst, not &dst (or perhaps make returns a pointer, so one is not rqd.)
		errGet3 := datastore.GetMulti(c, relevantKeys3, relevantRFQs3)
		if errGet3 != nil{
			c.Errorf("Relevant RFQs not retrieved")
		}
		//the iterator for table rows
		for i:=0; i<len(relevantRFQs3); i++{
			relevantRFQs3[i].Iterator = i+1
		}	
		
		
		//Now the dashboard.
		io.WriteString(w, PageTop)
        io.WriteString(w, "User <b>"+ usr.Username + "</b> logged in.  <a href=\"/signout\">Logout</a>")
        io.WriteString(w, RInsDashHead)
        io.WriteString(w, RInsDashTable)        
		t1 := template.New("reinsdash.html")
		t1  = template.Must(t1.ParseFiles("reinsdash.html"))
		//execute for each RFQ
		for _,r := range relevantRFQs1{
			err4 := t1.Execute(w,r)
			if err4 != nil{
				c.Errorf("Error executing template")		
			}
		}
		io.WriteString(w, "</table><div>")
		io.WriteString(w, "<h4>RFQs already responded.</h4>")
		
		io.WriteString(w, RInsDashTable)
		//t1 := template.New("reinsdash.html")
		//t1  = template.Must(t1.ParseFiles("reinsdash.html"))
		//execute for each RFQ
		for _,r := range relevantRFQs2{
			err4 := t1.Execute(w,r)
			if err4 != nil{
				c.Errorf("Error executing template")		
			}
		}
		io.WriteString(w, "</table><div>")
		
		io.WriteString(w, "<h4>RFQs marked closed by insurer/broker.</h4>")
		io.WriteString(w, "<p>Unless reactivated by insurer/broker, items are deleted one week after they are marked for closing.</p>")
		io.WriteString(w, RInsDashTable)
		//t1 := template.New("reinsdash.html")
		//t1  = template.Must(t1.ParseFiles("reinsdash.html"))
		//execute for each RFQ
		for _,r := range relevantRFQs3{
			err4 := t1.Execute(w,r)
			if err4 != nil{
				c.Errorf("Error executing template")		
			}
		}
		io.WriteString(w, "</table><div>")
}

const RInsDashHead =`
<p> <a class = "btn btn-success" href="/rinsdboard"  role="button">Refresh Page</a> </p>

<h4>RFQs received but not responded.</h4>

`

const RInsDashTable = `
<p>Click on an RFQ ID to view details, and respond. </p>
<div>
<table class = "table">
<tr>
<td><b>Item</b></td>
<td><b>RFQ ID</b></td>
<td><b>Requested by</b></td>
<td><b>Ins. Class</b></td>
<td><b>Ins. Type</b></td>
<td><b>Ins Value</b></td>
<td><b>Ins. Retention</b></td>
</tr>
`
/*
<p>Click on an RFQ ID to view details, and respond. </p>
<div>
<table class = "table">
<tr>
<td><b>Item</b></td>
<td><b>RFQ ID</b></td>
<td><b>Requested by</b></td>
<td><b>Client</b></td>
<td><b>Ins. Class</b></td>
<td><b>Ins. Type</b></td>
<td><b>Ins Value</b></td>
<td><b>Ins. Retention</b></td>
<td><b>Remarks</b></td>
</tr>
*/

//Function to go to review summary when done uploading files.
func RevRinsFile(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	//recall user login creds
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//recall Rins' version of current RFQ
	Rinsrfq, err2 := SessRecallRinsRFQ(w,r)	
	if err2 != nil{
		c.Errorf("error reading session values")
		return
	}
	
	t := template.New("RinsMainRFQ.html")
	t  = template.Must(t.ParseFiles("templates/RinsMainRFQ.html"))
	err = t.Execute(w,Rinsrfq)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}	
}




//Reinsurer's initial response to RFQ. Just saving at this point.
func RinsInitResp(w http.ResponseWriter, r *http.Request){
	
	c := appengine.NewContext(r)
	//recall user login creds
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//recall Rins' version of current RFQ
	Rinsrfq, err2 := SessRecallRinsRFQ(w,r)	
	if err2 != nil{
		c.Errorf("error reading session values")
		return
	}
	//recall the location in RinsList of the actual/complete RI's version for updating rfqKeyInfo.RfqKey
	RinsrfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
	// read the form
	r.ParseForm()
	
		
	if r.Form.Get("MaxCover") != "" {// && r.Form.Get("MaxCover") != Rinsrfq.Limit[0] {
		Rinsrfq.Limit = strArrAppZero(Rinsrfq.Limit,r.Form.Get("MaxCover"))
	}
	if r.Form.Get("Prem") != "" {//&& r.Form.Get("Prem") != Rinsrfq.Premium[0] {
		Rinsrfq.Premium = strArrAppZero(Rinsrfq.Premium,r.Form.Get("Prem"))
	}
	if r.Form.Get("Comm") != "" {//&& r.Form.Get("Comm") != Rinsrfq.Commission[0] {
		Rinsrfq.Commission = strArrAppZero(Rinsrfq.Commission, r.Form.Get("Comm"))
	}	
	if r.Form.Get("remarks") != ""{
		Rinsrfq.ConvAuth = strArrAppZero(Rinsrfq.ConvAuth, Rinsrfq.RIUsername)
		convArr := []string{r.Form.Get("remarks"), " ", "(",time.Now().Format("2 Jan 2006 15:04 UTC"),")"}
		convStr := strings.Join(convArr, "")	
		Rinsrfq.Conversation = strArrAppZero(Rinsrfq.Conversation, convStr)
	}	
	if r.Form.Get("private") != ""{
		Rinsrfq.Private = r.Form.Get("private") 
	} 	
	//Must autosave into datastore, but have to pull the relevant RFQ first and update it.
	//Do not push this partial version.
	var realRinsrfq RInsRFQList
	errGet := datastore.Get(c, RinsrfqKeyInfo.RfqKey, &realRinsrfq)
	if errGet != nil{
		c.Errorf("couldn't Get Rins RFQ from RinsRFQList in RinsInitResp")
		return
	}
	//make sure they match
	if realRinsrfq.RFQId != Rinsrfq.RFQId || realRinsrfq.RIUsername != Rinsrfq.RIUsername{
		c.Errorf("wrong RFQ in RinsInitResp")
		return	
	}
		
	realRinsrfq.Limit = Rinsrfq.Limit 
	realRinsrfq.Premium = Rinsrfq.Premium
	realRinsrfq.Commission = Rinsrfq.Commission
	realRinsrfq.ConvAuth = Rinsrfq.ConvAuth
	realRinsrfq.Conversation = Rinsrfq.Conversation
	realRinsrfq.Private = Rinsrfq.Private
	
	_, errPut := datastore.Put(c, RinsrfqKeyInfo.RfqKey, &realRinsrfq)
	if errPut != nil{
		c.Errorf("couldn't put updated Rins RFQ into RinsRFQList in RinsInitResp")
		return
	}
	
	//Must update session cookies as well
	session, err := store.Get(r,"sessRins")
        if err!=nil{
        	http.Error(w,err.Error(),500)
        	return
        }       
    session.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session.Values["CurrRinsRFQ"] = Rinsrfq
    session.Save(r,w)
    
    //then go back to the view
    t := template.New("RinsMainRFQ.html")
	t  = template.Must(t.ParseFiles("templates/RinsMainRFQ.html"))
	err = t.Execute(w,Rinsrfq)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}
return
}

//Function to actually send a response to insurer.
//This is clicking a button on the RFQ Mains page. (RinsMainRFQ)
//should add to response received in RFQ just once.
func RinsRespMsg(w http.ResponseWriter, r *http.Request){
//The elements of the response are already saved. 
//RI will also change its own resp status to 2 and insurer's to 3.
	c := appengine.NewContext(r)
	_,err0 := SignInRecall(w,r)
		if err0 != nil{
			c.Errorf("sign in error in RinsRespMsg")
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//recall Rins' version of current RFQ
	Rinsrfq, err2 := SessRecallRinsRFQ(w,r)	
	if err2 != nil{
		c.Errorf("error reading session values")
		return
	}
	//recall the location in RinsList of the actual/complete RI's version for updating rfqKeyInfo.RfqKey
	RinsrfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}	
	
	//recall insurer's RFQ and insert the status change.
	var x RFQ
	errRFQGet := datastore.Get(c, Rinsrfq.ReqRcvd, &x)
	if errRFQGet != nil{
		c.Errorf("could not Get ins RFQ in RinsRespMsg")
		return
	}
	//check that it is the same RFQ and that you have not responded already
	if x.RFQId == Rinsrfq.RFQId{
		//set the status to response rcvd.
		x.Status = 3		
		//check to see if responded earlier.
		strCheck := strings.Join(x.RespRcvdFrom, " ")
		if !strings.Contains(strCheck, Rinsrfq.RIUsername) {//have not responded previously		
			//update who the response is from		
			x.RespRcvdFrom = strArrAppZero(x.RespRcvdFrom, Rinsrfq.RIUsername)
			//and the key to the response in RinsRFQList
			x.RespRcvdKeys = keyArrAppZero(x.RespRcvdKeys,  RinsrfqKeyInfo.RfqKey)
			//upload the info
		}
		_, err := datastore.Put(c, Rinsrfq.ReqRcvd, &x)
		if err != nil {
			c.Errorf("could not Put updated ins RFQ back")
			return
		}
	}else{
		c.Errorf("got wrong ins RFQ")
		return
	}
	
	//set the response status as responded
	Rinsrfq.RespStatus = 2
	Rinsrfq.RespDate = time.Now().Format("2 Jan 2006 15:04 UTC")
	//store update in RinsRFQList
//NOTE: This implies data like client name and insurer's filenames, blobs etc will
//also get stored in rinsList. Stored but never used. 
//If this is undesirable, better to pull out the current version from base, 
//copy only necessary information and put back.
	_, errPut := datastore.Put(c, RinsrfqKeyInfo.RfqKey, &Rinsrfq)
	if errPut != nil{
		c.Errorf("could not put updated status to RinsRFQList in RinsRespMsg")
	}
	session, err := store.Get(r,"sessRins")
        if err!=nil{
        	http.Error(w,err.Error(),500)
        	return
        }       
    session.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session.Values["CurrRinsRFQ"] = Rinsrfq
    session.Save(r,w)
    
    //then go back to the view
    t := template.New("RinsMainRFQ.html")
	t  = template.Must(t.ParseFiles("templates/RinsMainRFQ.html"))
	err = t.Execute(w,Rinsrfq)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}
return
}