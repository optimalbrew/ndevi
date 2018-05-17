//Insurer user interface to display pages and options post login.
//TODO: function to delete RFQ (must call function to delete associated blobfiles)


package main

import(
	"net/http"
	//"fmt"
	"io"
	"sort"
		
	"strings"
	"strconv"
	"html/template"
	"appengine"
	"appengine/datastore"
	"appengine/blobstore"

	"time"	
	"github.com/gorilla/sessions"
	"github.com/gorilla/mux"
)

//Used in Dashboard for insurer staff. Displays new RFQ form.
func InsPage(w http.ResponseWriter, r *http.Request){	
		c := appengine.NewContext(r)
		//recall user from session data
		user,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
		
		if user.UserType == "RIns"{
			return		//only show to those looking for reinsurance.
		}
		
		rfq := new(RFQ)
		rfq.InsEmpID = user.Username 
		rfq.Treaty = 0 //facultative
		//Execute main template
		t := template.New("insMainRFQ.html")
		t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
		err = t.Execute(w,rfq)
		if err != nil{
		c.Errorf("Error executing template")
		return		
		}
}

//Same as above, but to display the Treaty RFQ form
func TreatyPage(w http.ResponseWriter, r *http.Request){	
		c := appengine.NewContext(r)
		//recall user from session data
		user,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
		
		if user.UserType == "RIns"{
			return		//only show to those looking for reinsurance.
		}
		
		rfq := new(RFQ)
		rfq.InsEmpID = user.Username 
		rfq.Treaty = 1 //indicates a Treaty RFQ
		//Execute main template
		t := template.New("insMainRFQ.html")
		t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
		err = t.Execute(w,rfq)
		if err != nil{
		c.Errorf("Error executing template")
		return		
		}
}


//Function to read RFQ values from forms. Username is filled in from session info.
//Show this as submit button on main template only if no RFQ is on record
func CreateRFQ(w http.ResponseWriter, r *http.Request){
		
		c := appengine.NewContext(r)
		//recall user		
		user,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
			}
		usr := user.Username		

		comp := strings.Split(usr,"@") //easier than querying datastore for company info
					
		//generate random RFQ ID. Idea is to use this to link all users associated with it.
		rfqID := RandStrings(10)
		//parse RFQ form
       	r.ParseForm()		
				
		RIemails := r.Form.Get("RIemails")
		RIemails = strings.ToLower(RIemails)
		RIemails = strings.Replace(RIemails, "<"," ",-1)
		RIemails = strings.Replace(RIemails, ">"," ",-1)
		RIemails = strings.Replace(RIemails, ","," ",-1) //replace comma by whitespace
		RIemailsSplit := strings.Fields(RIemails) //convert to []string	
		numRinsToAdd := len(RIemailsSplit)
	
		for i:=0;i<numRinsToAdd;i++{
			//check for duplicates in list to be added	
			dupCheck := strings.Count(RIemails, RIemailsSplit[i])
			if  dupCheck >1 {
				RIemails = strings.Replace(RIemails, RIemailsSplit[i]," ",(dupCheck-1))
			}	
		}	
		//Reconstruct the []string
		RIemailsSplit = strings.Fields(RIemails) //convert to []string	
		numRinsToAdd = len(RIemailsSplit)
	
		countValid := 0	//strings with @
		for i:=0;i<numRinsToAdd;i++{
			if 	strings.Contains(RIemailsSplit[i],"@"){			
				countValid++
			}
		}
		//"Valid" emails to add, must have @, spaces already removed 
		RinsAddValid := make([]string, countValid)
		l := 0 //some iterator
		for i := 0; i<numRinsToAdd;i++{
			if 	strings.Contains(RIemailsSplit[i],"@"){
				RinsAddValid[l] = RIemailsSplit[i]  
				l++
			}
		}
		//If unselected, the default is all bids, as method would be 0.
		method,errMethod := strconv.Atoi(r.Form.Get("method"))
		if errMethod != nil{
			c.Errorf("atoi problem in getting method")
		}
				
		rfq := RFQ{
			RFQId: rfqID,
			Insurer: comp[1],
			InsEmpID: usr,
			ClientName: r.Form.Get("client"),
			InsCat: r.Form.Get("cat"),
			NewCust: r.Form.Get("RadioProd"),
			InsType: r.Form.Get("RItype"),
			InsVal: r.Form.Get("sumIns"),
			InsDed: r.Form.Get("reten"),
			Brokerage: r.Form.Get("brokerage"),
			Period: r.Form.Get("period"),
			Method: method,
			EstPrem: r.Form.Get("estprem"),
			InsRemarks:r.Form.Get("revised.remarks"),
			InsPrivate:r.Form.Get("insprivate"),	
			ReInsList: RinsAddValid,	
			RFQDate: time.Now(),
			RFQStrTime: time.Now().Format("2 Jan 2006 15:04 UTC"), // time.Now(),	
			Status: 1,
			Modified: "new",		
		}				
	vars := mux.Vars(r)
		
	if vars["key"] == "Treaty"{
		rfq.Treaty = 1
	}
		
	//mark it as saved. If the put is not successful, it will exit. So session cookie 
	//will not be saved either	
	//store it in the database	
	keyRFQ := datastore.NewIncompleteKey(c, "RFQs", RFQKey(c))
    keyFromPut, err2 := datastore.Put(c, keyRFQ, &rfq)
    if err2 != nil {
        http.Error(w, err2.Error(), http.StatusInternalServerError)
        return
    }
	
	//Cookie for RFQ ID and its key
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
    cookieForLocRFQkey := sessLocKey{
						RFQId: rfq.RFQId,
						RfqKey: keyFromPut,
					}             
    session1.Values["Key"] = cookieForLocRFQkey
    		
	//Cookie with RFQ details
	session2, errSess := store.Get(r,"sessIns")
        if errSess!=nil{
        	http.Error(w,errSess.Error(),500)
        	return
        }       
    session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session2.Values["currRFQ"] = rfq        
    //save both cookies
    sessions.Save(r,w)  //note, not (w,r)
	
    //Now redirect to the files upload page. 
    http.Redirect(w, r, "/handleroot", http.StatusFound)
}


//Review newly created RFQ. 
func ReviewRFQ(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)	
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//Cookie RFQ
	rfqSess, err := SessRecallRFQ(w,r)	
	if err != nil{
		c.Errorf("error reading session values in ReviewRFQ")
		return
	}
	//Use cookie to recall the rfq's location in data store. The id is .RFQId, and the key is .RfqKey
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
	var rfq RFQ
	errGet := datastore.Get(c,rfqKeyInfo.RfqKey, &rfq)
	if errGet !=nil{
		c.Errorf("Couldn't Get RFQ in ReviewRFQ: %v", errGet)
		return
	}
	
	if rfq.RFQId != rfqSess.RFQId{
		c.Errorf("Session and Store Id mismatch in ReviewRFQ.")
		return
	}
	
	//Display review page
	if rfq.Status != 3 {				
		t := template.New("insMainRFQ.html")
		t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
		err = t.Execute(w,rfq)
		if err != nil{
			c.Errorf("Error executing template in ReviewRFQ")
			return		
		}
	}else{
		//recall the response from session
		resp, errResp := SessRecResp(w,r)
		if errResp != nil{
			c.Errorf("could not read sessResp cookie in pm")
			return
		}
		t := template.New("insMainStat3Top.html")
		t  = template.Must(t.ParseFiles("templates/insMainStat3Top.html"))
		err = t.Execute(w,rfq)
		if err != nil{
			c.Errorf("Error executing template top")
			return		
		}	
		t2 := template.New("insQuotes.html")
		t2  = template.Must(t2.ParseFiles("templates/insQuotes.html"))
		err = t2.Execute(w,resp)
		if err != nil{
			c.Errorf("Error executing template quote")
			return					
		}
	}	
}


//Share RFQ. Save RFQ to data store and send emails to RIs
//This should only show up on saved RFQs (review and reviewSaved)
func ShareRFQ(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)	
	//check log in status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//rfq key cookie
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
	//cookie RFQ
	rfq, err := SessRecallRFQ(w,r)	
	if err != nil{
		c.Errorf("error reading session values")
		return
	}
	
	if len(rfq.ReInsList) == 0 { //no one to share with!
		io.WriteString(w, PageTop)
		io.WriteString(w, "No reinsurers added! Add some reinsurers' email IDs first.")
		io.WriteString(w, PageBot)
		return
	}
	
    
    //Associate request with concerned RIs   
    errAssoc := AssocRFQKey(c, w, rfqKeyInfo.RfqKey, rfq)
    	if errAssoc !=nil{
    		c.Errorf("RFQ association error")
    	}
    //email to RIs: note third arg here is the full set not subset (e.g. adding a new RI)
    errmail := sendEmailRFQ(c, rfq, rfq.ReInsList) //initial RFQ email
    	if errmail != nil{
    		c.Errorf("sendemail error.")
	    }
	rfq.Status = 2 //to indicate saved and emailed (will only update if no error)			
	//save to datastore after successfully associating with reinsurers and emailing them.
	//save updated Status to datastore.
	_, err2 := datastore.Put(c, rfqKeyInfo.RfqKey, &rfq)
    if err2 != nil {
        http.Error(w, err2.Error(), http.StatusInternalServerError)
        return
    }    	
    session2, err3 := store.Get(r,"sessIns")
        if err3!=nil{
        	http.Error(w,err3.Error(),500)
        	return
        }       
    session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session2.Values["currRFQ"] = rfq
                
    session2.Save(r,w)  //note, not (w,r)
    	 
	http.Redirect(w,r,"/insdboard", http.StatusFound)	
}		



//Function to associate RFQ key to relevant RI's. 
//Should only show up for unshared RFQs?? It is okay as long as the rfq has the correct list of RIs
//and all of them need to look at this again.
func AssocRFQKey(c appengine.Context, w http.ResponseWriter, key *datastore.Key, rfq RFQ)(err2 error){	
	//search for current associations and delete them all.	
	
	if rfq.Status != 1{
		c.Errorf("cannot wipe the slate clean with RinsList, may delete conversations, files in AssocRFQKey")
		return
	}
	
	q := datastore.NewQuery("RInsList").Filter("RFQId =", rfq.RFQId)	
		var check RInsRFQList
		keyCheck, errCheck := q.KeysOnly().GetAll(c, &check)
		if errCheck != nil{
			err2 = errCheck
			c.Errorf("errCheck in assocRFQkey")
		}
	//delete all current associations irrespective of status
	if len(keyCheck)!=0{
		errDel := datastore.DeleteMulti(c, keyCheck)
		if errDel != nil{
			c.Errorf("Could not delete existing associations")
			err2 = errDel
		}	
	}
	//Now associate all current Rins with the RFQ: store a few things for querying	 
	for i:=0;i<len(rfq.ReInsList);i++{
		listItem :=	RInsRFQList{
						RIUsername: rfq.ReInsList[i],
						RFQId: rfq.RFQId,
						ReqRcvd: key, //from the function args
						RespStatus: 1, //1: response needed, 0 done.
						RFQDate: rfq.RFQDate,
						Insurer: rfq.Insurer,	
						InsEmpID: rfq.InsEmpID,	
						InsCat: rfq.InsCat,
					}			
		key1 := datastore.NewIncompleteKey(c, "RInsList", RInsListKey(c))
    	_, err := datastore.Put(c, key1, &listItem)
    	if err != nil {
         	http.Error(w, err.Error(), http.StatusInternalServerError)
           	c.Errorf("Could not associate RI")
        	err2 = err
            return
		}
	}
return
}


//helper function to check for and remove duplicates in string slices
func DupCheck(s []string) (sOut []string){
	sStr := strings.Join(s," ") //join with spaces
	numStr := len(s)
	for i:=0;i<numStr;i++{
		dup := strings.Count(sStr,s[i])
		if dup>1{
			sStr = strings.Replace(sStr, s[i], " ", dup-1)
		}
	}
	//convert back to []string
	sOut = strings.Fields(sStr)
return
}

//function to delte and add RI's to existing RFQ: Current RFQ and its key must be in session
func AddRins(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	//check log in status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	
	//Cookie RFQ
	rfq, err := SessRecallRFQ(w,r)	
	if err != nil{
		c.Errorf("error reading RFQ session values in addrins")
		return
	}
	//Use cookie to recall the rfq's location in data store. The id is .RFQId, and the key is .RfqKey
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
	
	resp := new(Resp) 
	var errResp error
	
	if rfq.Status == 3{ //will need this for displaying quotes.
		//cookie for (status 3, received responses)
		*resp, errResp = SessRecResp(w,r) //used *resp instead of just resp
		if errResp != nil{
			c.Errorf("could not read sessResp cookie in Addrins")
			return
		}	
	}
	
	r.ParseForm()
	//Rins to remove from list
	RInsDel := r.Form["RInsDel[]"]
	//numOrig := len(rfq.ReInsList)
	numRInsDel := len(RInsDel)	
	
			
	//remove these guys from RInsList for this rfq. Only for status 1, and 2. Else for stat 3, 
	//would have to remove from rinslist, and from RFQ, resprcvdfrom, keys, numresponses etc.	 
	if rfq.Status != 3 {	//|| rfq.Status == 3 
		for i:=0;i<numRInsDel;i++{
			var RInsListDel []RInsRFQList //these are the relevant RFQs
	  		q := datastore.NewQuery("RInsList").Filter("RIUsername =", RInsDel[i]).Filter("RFQId =", rfq.RFQId)//.Offset(10).Limit(10)           
        	//GetAll is a method for type query, and needs &dst, not dst. Unlike Get or GetMulti
        	keys, err := q.GetAll(c, &RInsListDel)
        	if err != nil{
        		c.Errorf("Query Error")
        	}
        	if len(keys)==0{ //reinsurer not associated yet.
        		break
        	}
        	//check for duplicates
        	if len(keys) >1{
        		c.Errorf("multiple matches in RinsList in addrins")
        	}
        	// should remove associated blobfiles as well, or they'll be orphaned.
        	for j := 0; j< len(keys);j++{
        		blobs := RInsListDel[j].RinsFileBlobs
        		if len(blobs)>0{	//safety wrap
        			errblobDel := blobstore.DeleteMulti(c,blobs)
        			if errblobDel != nil {
        				c.Errorf("could not delete associated blob files in addrins")
        			}
        		}
        	}
        	err2 := datastore.DeleteMulti(c,keys)
        	if err2 != nil{
        		c.Errorf("Could not delete from RinsList (in AddRins)")
        	}        	
    	}
    }

	//remove strings from rinsdel from the original	
	strOrig := strings.Join(rfq.ReInsList," ") //convert to single string with spaces
	for i:=0;i<numRInsDel;i++{ //should be 0 for status 3
		strOrig = strings.Replace(strOrig,RInsDel[i]," ",1) //delete the ones that match 
	}	
	
	//convert it back to []string
	ModRinsList := strings.Fields(strOrig) //using any whitespace as break
	numAftDel := len(ModRinsList)
	
	//prep the list of new additions.
	RIemails := r.Form.Get("RIemails")
	RIemails = strings.ToLower(RIemails)
	RIemails = strings.Replace(RIemails, "<"," ",-1)
	RIemails = strings.Replace(RIemails, ">"," ",-1)
	RIemails = strings.Replace(RIemails, ","," ",-1) //replace comma by whitespace
	RIemailsSplit := strings.Fields(RIemails) //convert to []string	
	numRinsToAdd := len(RIemailsSplit)
	
	//to prevent simultaneous deletion and addition in one step.
	strOrigAgain := strings.Join(rfq.ReInsList," ") 
	for i:=0;i<numRinsToAdd;i++{
		//check for duplicates in list to be added	
		dupCheck := strings.Count(RIemails, RIemailsSplit[i])
		//Don't allow deletion followed by immediate addition
		dupCheckOrig := strings.Count(strOrigAgain, RIemailsSplit[i])
		if  dupCheck >1 {
			RIemails = strings.Replace(RIemails, RIemailsSplit[i]," ",(dupCheck-1))
		}	
		if dupCheckOrig >0 {
			RIemails = strings.Replace(RIemails, RIemailsSplit[i]," ",-1)
		}
	}	
	//Reconstruct the []string
	RIemailsSplit = strings.Fields(RIemails) //convert to []string	
	numRinsToAdd = len(RIemailsSplit)
	
	countValid := 0	//strings with @
	for i:=0;i<numRinsToAdd;i++{
		if 	strings.Contains(RIemailsSplit[i],"@"){			
			countValid++
		}
	}
	//Valid emails to add 
	RinsAddValid := make([]string, countValid)
	l := 0 //some iterator
	for i := 0; i<numRinsToAdd;i++{
		if 	strings.Contains(RIemailsSplit[i],"@"){
			RinsAddValid[l] = RIemailsSplit[i]  
			l++
		}
	}  
	//revise actual number to add
	numRinsToAdd = countValid
	if numAftDel+numRinsToAdd == 0{
		rfq.Status = 1 //cannot be "shared" with zero reinsurers!
		rfq.ReInsList = make([]string,0)
		
	}else{		
		tempList := make([]string,numAftDel+numRinsToAdd)	
		//Now to fill in templist. First copy the old ones, provided they are not in DelList	
		//fill in the old ones (post deletion) first 		
		for i := 0;i< numAftDel;i++{
			tempList[i] = ModRinsList[i]//duplication checked.
		}			
		//fill in the new folks
		for i := 0;i< numRinsToAdd;i++{
			tempList[numAftDel+i] = RinsAddValid[i] //duplication checked
		}
		//now update the rfq list
		rfq.ReInsList = tempList	
	}		
	
	
	//save the rfq in datastore, and session info.
    _, err2 := datastore.Put(c, rfqKeyInfo.RfqKey, &rfq)
    if err2 != nil {
        http.Error(w, err2.Error(), http.StatusInternalServerError)
        return
    }
    //save session   
    session2, err3 := store.Get(r,"sessIns")
        if err3!=nil{
        	http.Error(w,err3.Error(),500)
        	return
        }       
    session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session2.Values["currRFQ"] = rfq
                
    session2.Save(r,w)  //note, not (w,r)
		//redirect back to review (as per status)
		 
    if rfq.Status == 1{    //Saved but not sent.
		t := template.New("insMainRFQ.html")
		t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
		err = t.Execute(w,rfq)
		if err != nil{
		c.Errorf("Error executing template")
		return		
		}
	} else if rfq.Status == 2 || rfq.Status == 3 { //already shared with original reinsurers. Removed deleted ones as well.
		//update RinsList struct with the new guys (do not use associate) may change status for some.			 
			for i:=0;i<numRinsToAdd;i++{
				listItem :=	RInsRFQList{
						RIUsername: RinsAddValid[i], //not the entire rfq.ReInsList[i],
						RFQId: rfq.RFQId,
						ReqRcvd: rfqKeyInfo.RfqKey, //from the function args
						RespStatus: 1, //1: response needed, 0 done.
						RFQDate: rfq.RFQDate,
						Insurer: rfq.Insurer,	
						InsEmpID: rfq.InsEmpID,	
						InsCat: rfq.InsCat,
					}			
				key1 := datastore.NewIncompleteKey(c, "RInsList", RInsListKey(c))
    			_, err := datastore.Put(c, key1, &listItem)
    			if err != nil {
         			http.Error(w, err.Error(), http.StatusInternalServerError)
           			c.Errorf("Could not associate RI")
            		return
				}
			}
	
		//email the new guys about the rfq
		 errmail := sendEmailRFQ(c, rfq, RinsAddValid) //initial RFQ email
    		if errmail != nil{
    			c.Errorf("sendemail error.")
	    	}
		//now go back to the review.
		if rfq.Status == 2 {
			t := template.New("insMainRFQ.html")
			t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
			err = t.Execute(w,rfq)
			if err != nil{
				c.Errorf("Error executing template")
			return		
			}
		
		
		}else if rfq.Status == 3{ //status must be 3
			//run the template
			t := template.New("insMainStat3Top.html")
			t  = template.Must(t.ParseFiles("templates/insMainStat3Top.html"))
			err = t.Execute(w,rfq)
			if err != nil{
				c.Errorf("Error executing template top")
				return		
			}	
			t2 := template.New("insQuotes.html")
			t2  = template.Must(t2.ParseFiles("templates/insQuotes.html"))
			err = t2.Execute(w,resp)
			if err != nil{
				c.Errorf("Error executing template quote")
				return		
			}
		}
	}			
return
}


//Function to (locate and display) go to RFQ from dashboard RFQID link
func InsLocRFQ(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	usr, err := SignInRecall(w,r)	
	if err != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		c.Errorf("error reading session values")
		return
	}
		
	vars := mux.Vars(r)
	rfqId := vars["key"]
	
	//Without global info on IDs and Keys, have to search for it.
	//query to locate the RFQ and to ensure it matches both the rfq ID and the user's ID
	var insFindrfq []RFQ
	q := datastore.NewQuery("RFQs").Filter("InsEmpID =",usr.Username).Filter("RFQId =", rfqId)           
        locKey, err2 := q.GetAll(c, &insFindrfq)
        if err2 != nil{
        	c.Errorf("RFQ not found")
        	return
        }	
    if len(insFindrfq)>1{
    	c.Errorf("More than one match (InsLockRFQ).") //should be at most one result
    	return
    }
    rfq := insFindrfq[0]
    rfqKey :=locKey[0] 
    //session cookie for the located RFQ's key
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
    cookieForLocRFQkey := sessLocKey{
						RFQId: rfq.RFQId,
						RfqKey: rfqKey,
					}             
    session1.Values["Key"] = cookieForLocRFQkey
    
    //session1.Save(r,w)
    //session cookie for the located RFQ itself
       
    session2, err3 := store.Get(r,"sessIns")
        if err3!=nil{
        	http.Error(w,err3.Error(),500)
        	return
        }       
    session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session2.Values["currRFQ"] = rfq
                
    
      
    if rfq.Status == 1 || rfq.Status == 2 {
    	sessions.Save(r,w)  //note, not (w,r) 
		t := template.New("insMainRFQ.html")
		t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
		err = t.Execute(w,rfq)
		if err != nil{
		c.Errorf("Error executing template")
		return		
		}
	} else if rfq.Status == 5 || rfq.Status == 6 { //deleted unshared
		sessions.Save(r,w)  //note, not (w,r)
		t := template.New("reviewDel.html")
		t  = template.Must(t.ParseFiles("reviewDel.html"))
		err = t.Execute(w,rfq)
		if err != nil{
		c.Errorf("Error executing template")
		return		
		}	
	}else if rfq.Status == 3{//received responses, this is going to be more involved.
		//find who's responded and grab their RinsRFQLists
		respKeys := rfq.RespRcvdKeys
		numresp := len(respKeys)
		
		//type RinsRespListSort []RInsRFQList defined in helpers.go
		
		RinsResp := make(RinsRespListSort, numresp)
		errGetRins := datastore.GetMulti(c, respKeys, RinsResp)
		if errGetRins != nil{
			c.Errorf("could not Get rins responses in inslocRFQ")
		}
		
		//verfiy these are associated with the current RFQ
		for i := 0; i< numresp; i++{
			if RinsResp[i].RFQId != rfq.RFQId{
				c.Errorf("resp RFQ from RinsList does not match current RFQ in inslocRFQ status 3")
				return
			}	
		}
		//if verified, then build the struct to pass to the template.
		//but before that, sort the array according to latest premium
		sort.Sort(RinsResp)
		
		premStr := make([]string, numresp) //store latest premium for median calculation
		premFloat := make([]float64, 0, numresp) //3rd arg is optional cap, for median calc.

		for i := 0; i< numresp;i++{
			premStr[i] = RinsResp[i].Premium[0]  
		}
		resp := new(Resp)
		resp.RowNames = []string{"Max Cover", "Premium", "Commission", "Files", "Conversation"}
		rins := make([]string, numresp)
		rinskey := make([]*datastore.Key, numresp)
		limit := make([][]string, numresp)
		premium := make([][]string, numresp)
		comm := make([][]string, numresp)
		files := make([][]string, numresp)
		auth := make([][]string, numresp)
		msg := make([][]string, numresp)
		
		//Calculate the median
		for i := 0; i< numresp;i++{
			if premStr[i] != ""{
				premIfloat, errI := strconv.ParseFloat(premStr[i],64)
				if errI != nil{
					c.Errorf(premStr[i])
					c.Errorf("Problem converting string Premium to float in InsLocRfq")
				}
				premFloat = append(premFloat,premIfloat)
			}
		}
		if len(premFloat)>2{
			floatMedian, errMedian := median(premFloat)
			//median is not calculated for less than 3 responses.
			if errMedian == nil{
				resp.Median = strconv.FormatFloat(floatMedian,'f',2,64)
			}
		}
		//Now hide the lowest bid
		if rfq.Method == 1{ //second price
			RinsResp[0].Premium = []string{"Hidden by system"}
			//modify later to incorporate commission, other calculation.
		}		
		//when calling the struct, call the categories in the end so they are in the 0th position

		
				
		for i := 0; i< numresp;i++{
			rins[i] = RinsResp[i].RIUsername
			rinskey[i]= respKeys[i]
			limit[i] = RinsResp[i].Limit				
			premium[i] = RinsResp[i].Premium
			comm[i] = RinsResp[i].Commission
			files[i] = RinsResp[i].RinsFileNames
			auth[i] = RinsResp[i].ConvAuth
			msg[i] = RinsResp[i].Conversation
			premStr[i] = RinsResp[i].Premium[0]  
		}	
		resp.Rins = rins
		resp.Respkey = rinskey
		resp.Limit = limit
		resp.Premium = premium
		resp.Commission = comm
		resp.Files = files
		resp.ConvAuth = auth
		resp.Conversation = msg	
		
		
		//save the response to session. Will need it to look
		//at the RFQ once we insurer sends a pm to a reinsurer
		sessionResp, errResp := store.Get(r,"sessResp")
        if errResp!=nil{
        	http.Error(w,errResp.Error(),500)
        	return
        }       
	    sessionResp.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
 	   //set session value        
    	sessionResp.Values["CurrResp"] = resp
		sessions.Save(r,w)  //note, not (w,r)
		
		//run the template
		t := template.New("insMainStat3Top.html")
		t  = template.Must(t.ParseFiles("templates/insMainStat3Top.html"))
		err = t.Execute(w,rfq)
		if err != nil{
			c.Errorf("Error executing template top")
			return		
		}
		//This will depend on method and also layout choices.
		t2 := template.New("insQuotes.html")
		t2  = template.Must(t2.ParseFiles("templates/insQuotes.html"))
		err = t2.Execute(w,resp)
		if err != nil{
			c.Errorf("Error executing template quote")
			return		
		}
	}		
}

//insurer's personal message to some reinsurer via Stat 3 (response recvd) conversation
//add the response to session values so it gets updated
//add the response to datastore (write to convAuth and conversation. Change resp stat to 1.
func pm(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	resp, errResp := SessRecResp(w,r)
	if errResp != nil{
		c.Errorf("could not read sessResp cookie in pm")
		return
	}
	//RFQ cookie
	rfq, errRFQ := SessRecallRFQ(w,r)	
	if errRFQ != nil{
		c.Errorf("error reading RFQ session values in pm")
		return
	}
	vars := mux.Vars(r)
	rinsToPM := vars["key"]
	
	rinsToRinsKey := make(map[string]*datastore.Key)
	for i := 0; i<len(resp.Rins); i++{
		rinsToRinsKey[resp.Rins[i]] = resp.Respkey[i]
	}
	
	//Key of RIns to pm
	KeyRinsToPM := rinsToRinsKey[rinsToPM]
	//grab the RinsRFQ using the key and make sure it is for the intended recipient and RFQ
	var x RInsRFQList
	errGet := datastore.Get(c, KeyRinsToPM, &x)
	if errGet != nil{
		c.Errorf("could not get Rins RFQ from key in pm")
		return
	}
	//check it
	if rfq.RFQId != x.RFQId || x.RIUsername != rinsToPM{
		c.Errorf("incorrect Rins RFQ or RIns to PM in pm")
		return
	}
	
	//now read the personal message
	r.ParseForm()
	msg := r.Form.Get("pm")
	auth := rfq.InsEmpID
	
	//Update conversation for Rins in RinsRFQList
	if msg != ""{
		x.ConvAuth = strArrAppZero(x.ConvAuth, auth)
		convArr := []string{msg, " ", "(",time.Now().Format("2 Jan 2006 15:04 UTC"),")"}
		convStr := strings.Join(convArr, "")	
		x.Conversation = strArrAppZero(x.Conversation, convStr)
		x.RespStatus = 1
		
		_, errPut := datastore.Put(c, KeyRinsToPM, &x)		
		if errPut != nil{
			c.Errorf("could not Put Rins RFQ with conv and stat update in pm")
			return
		}
		
		//find which rins was associated, and update the resp for that one.
		//This is to display through the template for now.
		for i := 0; i <len(resp.Rins); i++{
			if resp.Rins[i] == rinsToPM{
			resp.ConvAuth[i] = x.ConvAuth
			resp.Conversation[i] = x.Conversation
			break
			}
		}
		
		//refesh the page/template
	
	//save the response to session. Will need it to look
		//at the RFQ once we insurer sends a pm to a reinsurer
		sessionResp, errResp := store.Get(r,"sessResp")
        if errResp!=nil{
        	http.Error(w,errResp.Error(),500)
        	return
        }       
	    sessionResp.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
 	   //set session value        
    	sessionResp.Values["CurrResp"] = resp
		sessions.Save(r,w)  //note, not (w,r)
		
		//run the template
		t := template.New("insMainStat3Top.html")
		t  = template.Must(t.ParseFiles("templates/insMainStat3Top.html"))
		err := t.Execute(w,rfq)
		if err != nil{
			c.Errorf("Error executing template top")
			return		
		}	
		t2 := template.New("insQuotes.html")
		t2  = template.Must(t2.ParseFiles("templates/insQuotes.html"))
		err = t2.Execute(w,resp)
		if err != nil{
			c.Errorf("Error executing template quote")
			return		
		}
	}	
return
}



//insurer dashboard: See template examples on Go Docs
func InsurerDashboard(w http.ResponseWriter, r *http.Request){
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
		
	if usr.UserType == "RIns"{ //Rins cannot view insdboard
		c.Errorf("Rins cannot access Ins dboard")
		io.WriteString(w, PageTop)
		io.WriteString(w, "Invalid user type.")
		io.WriteString(w, PageBot)
		return
	}
	//search option, case sensitive has to be made lowercase
	//SearchBase(c, w, usr, "SKy")
	//Run queries to find the user's associated RFQ's with status code 1 (saved but not sent)
		var x1 []RFQ //these are the relevant RFQs with status 1
	  	q1 := datastore.NewQuery("RFQs").Filter("InsEmpID =",usr.Username).Filter("Status =", 1).Order("-RFQDate")           
        Keys1, err1 := q1.KeysOnly().GetAll(c, &x1)
        if err1 != nil{
        	c.Errorf("Status 1 Query Error")
        }      
        //Grab latest information through Get (Get Multi Doesn't work)
    	numKeys1 := len(Keys1)
       	insRFQStat1 := make([]RFQ,numKeys1)
       	for i := 0;i<numKeys1;i++{
       		 errGet := datastore.Get(c, Keys1[i], &insRFQStat1[i])
       		if errGet != nil{
        		http.Error(w, errGet.Error(),500)
        		c.Errorf("Get Multi failed Status 1 for insdboard")
        		return
        	}
       	}
        //set iterator for table rows.        
        for i := 0;i<len(insRFQStat1);i++{
        	insRFQStat1[i].Iterator = i+1
        }
	
	//find the ones that have already been shared
		var x2 []RFQ //these are the relevant RFQs with status 2
	  	q2 := datastore.NewQuery("RFQs").Filter("InsEmpID =",usr.Username, ).Filter("Status =", 2).Order("-RFQDate")           
        Keys2, err2 := q2.KeysOnly().GetAll(c, &x2)
        if err2 != nil{
        	c.Errorf("Status 2 Query Error")
        }
        //Grab latest information through Get (Get Multi Doesn't work)
    	numKeys2 := len(Keys2)
       	insRFQStat2 := make([]RFQ,numKeys2)
       	for i := 0;i<numKeys2;i++{
       		 errGet := datastore.Get(c, Keys2[i], &insRFQStat2[i])
       		if errGet != nil{
        		http.Error(w, errGet.Error(),500)
        		c.Errorf("Get Multi failed Status 2 for insdboard")
        		return
        	}
       	}     
        //set iterator for table rows.
        for i := 0;i<len(insRFQStat2);i++{
        	insRFQStat2[i].Iterator = i+1
        	}  
		
	//Find those with response received
		var x3 []RFQ //these are the relevant RFQs with status 3
	  	q3 := datastore.NewQuery("RFQs").Filter("InsEmpID =",usr.Username, ).Filter("Status =", 3).Order("-RFQDate")           
        Keys3, err3 := q3.KeysOnly().GetAll(c, &x3)
        if err3 != nil{
        	c.Errorf("Status 3 Query Error")
        }
        //Grab latest information through Get (Get Multi Doesn't work)
    	numKeys3 := len(Keys3)
       	insRFQStat3 := make([]RFQ,numKeys3)
       	for i := 0;i<numKeys3;i++{
       		 errGet := datastore.Get(c, Keys3[i], &insRFQStat3[i])
       		if errGet != nil{
        		http.Error(w, errGet.Error(),500)
        		c.Errorf("Get Multi failed Status 2 for insdboard")
        		return
        	}
       	}     
        //set iterator for table rows.
        for i := 0;i<len(insRFQStat3);i++{
        	insRFQStat3[i].Iterator = i+1
        	}  
	
        
	//Run queries to find RFQ's marked for deletion with status code 5
		var x5 []RFQ //these are the relevant RFQs with status 5
	  	q5 := datastore.NewQuery("RFQs").Filter("InsEmpID =",usr.Username).Filter("Status =", 5).Order("-RFQDate")           
        Keys5, err5 := q5.KeysOnly().GetAll(c, &x5)
        if err5 != nil{
        	c.Errorf("Status 5 Query Error")
        }      
        //Grab latest information through Get (Get Multi Doesn't work)
    	numKeys5 := len(Keys5)
       	insRFQStat5 := make([]RFQ,numKeys5)
       	for i := 0;i<numKeys5;i++{
       		 errGet := datastore.Get(c, Keys5[i], &insRFQStat5[i])
       		if errGet != nil{
        		http.Error(w, errGet.Error(),500)
        		c.Errorf("Get Multi failed Status 5 for insdboard")
        		return
        	}
       	}
        //set iterator for table rows.        
        for i := 0;i<len(insRFQStat5);i++{
        	insRFQStat5[i].Iterator = i+1
        }
                
        //Run queries to find RFQ's marked for deletion with status code 6
		var x6 []RFQ //these are the relevant RFQs with status 5
	  	q6 := datastore.NewQuery("RFQs").Filter("InsEmpID =",usr.Username).Filter("Status =", 6).Order("-RFQDate")           
        Keys6, err6 := q6.KeysOnly().GetAll(c, &x6)
        if err6 != nil{
        	c.Errorf("Status 6 Query Error")
        }      
        //Grab latest information through Get (Get Multi Doesn't work)
    	numKeys6 := len(Keys6)
       	insRFQStat6 := make([]RFQ,numKeys6)
       	for i := 0;i<numKeys6;i++{
       		 errGet := datastore.Get(c, Keys6[i], &insRFQStat6[i])
       		if errGet != nil{
        		http.Error(w, errGet.Error(),500)
        		c.Errorf("Get Multi failed Status 6 for insdboard")
        		return
        	}
       	}
        //set iterator for table rows.        
        for i := 0;i<len(insRFQStat6);i++{
        	insRFQStat6[i].Iterator = i+1
        }
        
        //To display the results, they all use the same template for now.
        //start with the ones that have not been shared yet
        io.WriteString(w, PageTop)
        io.WriteString(w, "<p>User <b>"+ usr.Username + "</b> logged in.  <a href=\"/signout\">Logout</a></p>")
        io.WriteString(w, insDashTop)
    	io.WriteString(w, insDashTable1)
                
		t := template.New("insdash.html")
		t  = template.Must(t.ParseFiles("insdash.html"))
		//execute for each RFQ
		for _,r := range insRFQStat1{
			err3 := t.Execute(w,r)
			if err3 != nil{
				c.Errorf("Error executing template table 1")		
			}
		}
		io.WriteString(w, "</table><div>")
		
		//now the ones with responses.
		
    	io.WriteString(w, "<h4>RFQs with responses</h4>")
        io.WriteString(w, "<p>Click on an RFQ ID to review, edit and send.</p>" + insDashTable1)        
		//t2 := template.New("insdash.html")
		//t2  = template.Must(t2.ParseFiles("insdash.html"))
		//execute for each RFQ
		for _,r := range insRFQStat3{
			errTemp := t.Execute(w,r)
			if errTemp != nil{
				c.Errorf("Error executing template table 2")		
			}
		}
		io.WriteString(w, "</table><div>")
		
		//now the ones that have been sent.
		
    	io.WriteString(w, "<h4>RFQs sent (no response yet)</h4>")
        io.WriteString(w, "<p>Click on an RFQ ID to review, edit and send.</p>" + insDashTable1)        
		//t2 := template.New("insdash.html")
		//t2  = template.Must(t2.ParseFiles("insdash.html"))
		//execute for each RFQ
		for _,r := range insRFQStat2{
			errTemp := t.Execute(w,r)
			if errTemp != nil{
				c.Errorf("Error executing template table 2")		
			}
		}
		io.WriteString(w, "</table><div>")
		
		//now the ones marked for deletion (shared).		
    	io.WriteString(w, "<h4>Closed RFQs</h4>")
    	io.WriteString(w, "<h5>Previously shared with reinsurers</h5>")
        io.WriteString(w, "<p>Items are deleted one week after they are marked for closing. Click on an RFQ ID to view and recover.</p>" + insDashTable2)        
		t2 := template.New("insdashClosed.html")
		t2  = template.Must(t2.ParseFiles("insdashClosed.html"))
		for i :=0 ; i<len(insRFQStat5); i++{
			errTemp := t2.Execute(w,insRFQStat5[i])
			if errTemp != nil{
				c.Errorf("Error executing template table 2")		
			}
			days := timeSinceDays(insRFQStat5[i].RFQDate)
			io.WriteString(w,"<td>"+strconv.Itoa(days)+" days</td></tr>")
		}
		
		io.WriteString(w, "</table><div>")
		
		//now the ones marked for deletion (unshared).		
    	io.WriteString(w, "<h5>Closed but never shared</h5>")
        io.WriteString(w, "<p>Items are deleted one week after they are marked for closing. Click on an RFQ ID to view and recover.</p>" + insDashTable1)        
		//t6 := template.New("insdash.html")
		//t6  = template.Must(t6.ParseFiles("insdash.html"))
		//execute for each RFQ
		for i :=0 ; i<len(insRFQStat6); i++{
			errTemp := t2.Execute(w,insRFQStat6[i])
			if errTemp != nil{
				c.Errorf("Error executing template table 2")		
			}
			days := timeSinceDays(insRFQStat6[i].RFQDate)
			io.WriteString(w,"<td>"+strconv.Itoa(days)+" days</td></tr>")
		}
		io.WriteString(w, "</table><div>")
		
		io.WriteString(w, PageBot)		
}

const insDashTop = `


<form class="navbar-form navbar-left" action = "/search" method="post">
  <a class = "btn btn-primary" href="/inspage"  role="button">New Facultative RFQ</a> 
    <a class = "btn btn-primary" href="/treaty"  role="button">New Treaty RFQ</a> 
  <a class = "btn btn-success" href="/insdboard"  role="button">Refresh Page</a> 
  <div class="form-group">
    <input type="text" name="query" class="form-control" placeholder="Search">
  </div>
  <button type="submit" class="btn btn-info">Search</button>
</form>

<br></br>
<h4>RFQs created but not sent</h4>
<p>Click on an RFQ ID to view details.</p>
`

const insDashTable1 = `
<div>
<table class = "table">
<tr>
<td><b>Item </b></td>
<td><b>RFQ ID </b></td>
<td><b>Client</b></td>
<td><b>Ins. Class</b></td>
<td><b>Ins Value</b></td>
<td><b>Ins. Retention</b></td>
</tr>
`

//to show days since marked closed
const insDashTable2 = `
<div>
<table class = "table">
<tr>
<td><b>Item </b></td>
<td><b>RFQ ID </b></td>
<td><b>Client</b></td>
<td><b>Ins. Class</b></td>
<td><b>Days since closing </b></td>
</tr>
`


//function to display search results
//this can have the same form as the dashboard itself.


//This one is no longer used. Hide it in the main.go file.
//function to show a revision form to edit existing RFQ
func ModifyRFQ(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	//session values needed to fill template
	//check log in status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//cookie RFQ
	rfq, err := SessRecallRFQ(w,r)	
	if err != nil{
		c.Errorf("error reading session values")
		return
	}	
	t := template.New("modifyRFQ.html")
	t  = template.Must(t.ParseFiles("modifyRFQ.html"))
	err = t.Execute(w,rfq)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}			
}

//Revising RFQ is simple from a newish form. But for Status 2/3, something must be conveyed
func ReviseRFQ(w http.ResponseWriter, r *http.Request){
	//Grab current RFQ information and Key location from cookies
	c := appengine.NewContext(r)	
	//check login status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	
	//rfq key cookie rfqKeyInfo.RfqKey
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
	//cookie RFQ
	rfq, err := SessRecallRFQ(w,r)	
	if err != nil{
		c.Errorf("error reading session values")
		return
	}
	// read in the new values
		r.ParseForm()				
		
		//now update the rfq with modified values
		rfq.RFQDate = time.Now()
		rfq.RFQStrTime = time.Now().Format("2 Jan 2006 15:04 UTC")
		if rfq.Status == 2 || rfq.Status == 3{ 
				rfq.Modified = "modified"
		}
	if r.Form.Get("insprivate") != ""{
			rfq.InsPrivate = r.Form.Get("insprivate")
		}	
	
	
	method,errMethod := strconv.Atoi(r.Form.Get("method"))
		if errMethod != nil{
			c.Errorf("atoi problem in getting method")
		}
		
	if rfq.Status == 1  { //not shared with Rins, no issues, just go ahead and change it
					
		if r.Form.Get("client") != ""{
			rfq.ClientName = r.Form.Get("client")
		}
		if r.Form.Get("cat") != ""{
			rfq.InsCat = r.Form.Get("cat")
		}	
		if r.Form.Get("RadioProd") != ""{
			rfq.NewCust = r.Form.Get("RadioProd")
		}		
		if r.Form.Get("RItype") != ""{
			rfq.InsType = r.Form.Get("RItype")
		}		
		if r.Form.Get("sumIns") != ""{
			rfq.InsVal = r.Form.Get("sumIns")
		}		
		if r.Form.Get("reten") != ""{
			rfq.InsDed = r.Form.Get("reten")
		}
		if r.Form.Get("brokerage") != ""{
			rfq.Brokerage = r.Form.Get("brokerage")
		}
		if r.Form.Get("estprem") != ""{
			rfq.EstPrem = r.Form.Get("estprem")
		}
		if r.Form.Get("period") != ""{
			rfq.Period = r.Form.Get("period")
		}
		if r.Form.Get("method") != ""{
			rfq.Method = method
		}
		if r.Form.Get("revised.remarks") != ""{
			rfq.InsRemarks = r.Form.Get("revised.remarks")
		}
				
	} else if rfq.Status == 2 || rfq.Status == 3 ||rfq.Status == 4{ //shared with Rins. Must inform of changes.
		if r.Form.Get("client") != ""{
			var str string = strings.Replace(rfq.ClientName, "Change history (revised entries): ","", 1)
			rfq.ClientName = string(r.Form.Get("client")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}
		if r.Form.Get("cat") != ""{
			var str string = strings.Replace(rfq.InsCat, "Change history (revised entries): ","", 1)
			rfq.InsCat = string(r.Form.Get("cat")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}	
		if r.Form.Get("RadioProd") != ""{
			var str string = strings.Replace(rfq.NewCust, "Change history (revised entries): ","", 1)
			rfq.NewCust = string(r.Form.Get("RadioProd")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}		
		if r.Form.Get("RItype") != ""{
			var str string = strings.Replace(rfq.InsType, "Change history (revised entries): ","", 1)
			rfq.InsType = string(r.Form.Get("RItype")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}		
		if r.Form.Get("sumIns") != ""{
			var str string = strings.Replace(rfq.InsVal, "Change history (revised entries): ","", 1)
			rfq.InsVal = string(r.Form.Get("sumIns")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}		
		if r.Form.Get("reten") != ""{
			var str string = strings.Replace(rfq.InsDed, "Change history (revised entries): ","", 1)
			rfq.InsDed = string(r.Form.Get("reten")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}
		if r.Form.Get("brokerage") != ""{
			var str string = strings.Replace(rfq.Brokerage, "Change history (revised entries): ","", 1)
			rfq.Brokerage = string(r.Form.Get("brokerage")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}
		if r.Form.Get("period") != ""{
			var str string = strings.Replace(rfq.Period, "Change history (revised entries): ","", 1)
			rfq.Period = string(r.Form.Get("period")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}
		if r.Form.Get("estprem") != ""{
			var str string = strings.Replace(rfq.EstPrem, "Change history (revised entries): ","", 1)
			rfq.EstPrem = string(r.Form.Get("estprem")+ " Change history (revised entries): "+ str + " ("+ rfq.RFQStrTime+")")
		}
		if r.Form.Get("revised.remarks") != ""{
			//var str string = strings.Replace(rfq.InsRemarks, "Change history (deleted): ","", 1)
			rfq.InsRemarks = string("(Revised: "+ rfq.RFQStrTime+") " + r.Form.Get("revised.remarks") )
		}
		
	}
	//save updated Status to datastore.
	_, errPut := datastore.Put(c, rfqKeyInfo.RfqKey, &rfq)
    if errPut != nil {
        http.Error(w, errPut.Error(), http.StatusInternalServerError)
        return
    } 
    //update session cookie for RFQ
    session2, errSess := store.Get(r,"sessIns")
        if errSess!=nil{
        	http.Error(w,errSess.Error(),500)
        	return
        }       
    session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session2.Values["currRFQ"] = rfq
                
    session2.Save(r,w)  //note, not (w,r)
	
	//If revised remarks: could also make them part of the conversation and store it in
	//RinsList entities as a broadcast.
	if r.Form.Get("revised.remarks") != ""{
		//pull all associated Rinslist entities.
		q := datastore.NewQuery("RInsList").Filter("RFQId =", rfq.RFQId)	
		var x []RInsRFQList
		keyCheck, errCheck := q.GetAll(c, &x)
		if errCheck != nil{
			c.Errorf("errCheck in reviseRFQ")
		}
		if len(keyCheck)>0{
				auth := rfq.InsEmpID
				msg := r.Form.Get("revised.remarks")
			for i := 0; i<len(keyCheck);i++{
				x[i].ConvAuth = strArrAppZero(x[i].ConvAuth, auth)
				convArr := []string{"[Broadcast message:] ", msg, " ", "(",time.Now().Format("2 Jan 2006 15:04 UTC"),")"}
				convStr := strings.Join(convArr, "")	
				x[i].Conversation = strArrAppZero(x[i].Conversation, convStr)
				x[i].RespStatus = 1
				_, errPut := datastore.Put(c, keyCheck[i], &x[i])		
				if errPut != nil{
					c.Errorf("could not Put Rins RFQ with conv and stat update in reviseRFQ")
					return
				}
			}								
		}
	}
	if rfq.Status == 1 || rfq.Status == 2{
		t := template.New("insMainRFQ.html")
		t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
		err = t.Execute(w,rfq)		
		if err != nil{
			c.Errorf("Error executing template")			
			return		
		}
	}else if rfq.Status == 3{
		//cookie for (status 3, received responses)
		resp, errResp := SessRecResp(w,r)
		if errResp != nil{
			c.Errorf("could not read sessResp cookie in pm")
			return
		}
		t := template.New("insMainStat3Top.html")
		t  = template.Must(t.ParseFiles("templates/insMainStat3Top.html"))
		err = t.Execute(w,rfq)
		if err != nil{
			c.Errorf("Error executing template top")
			return		
		}	
		t2 := template.New("insQuotes.html")
		t2  = template.Must(t2.ParseFiles("templates/insQuotes.html"))
		err = t2.Execute(w,resp)
		if err != nil{
			c.Errorf("Error executing template quote")
			return					
		}
	}		
return
}


		

//function to send email to all post modification
func ModAlert(w http.ResponseWriter, r *http.Request){
	//Grab current RFQ information and Key location from cookies
	c := appengine.NewContext(r)	
	//check log in status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//rfq key cookie rfqKeyInfo.RfqKey
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
	//cookie RFQ
	rfq, err := SessRecallRFQ(w,r)	
	if err != nil{
		c.Errorf("error reading session values")
		return
	}
	
	if len(rfq.ReInsList) == 0 { //no one to share with!
		io.WriteString(w, PageTop)
		io.WriteString(w, "No reinsurers added! Please add some reinsurers' email IDs first. ")
		io.WriteString(w,`Go back to the RFQ or <a class = "btn btn-primary" href="/insdboard"  role="button">Back to Dashboard</a></p>`)
		io.WriteString(w, PageBot)
		return
	}
	
	if rfq.Modified != "modified"{ //"new" or "modification emailed"
		io.WriteString(w,PageTop)
		io.WriteString(w,"<p>No new modifications detected since last email.")
		io.WriteString(w," <a class = \"btn btn-primary\" href=\"/insdboard\"  role=\"button\">Back to Dashboard</a></p>")
		io.WriteString(w,PageBot)
		return
	}
	
	//email to RIs: note third arg here is the full set not subset (e.g. adding a new RI)
    errmail := sendEmailRFQ(c, rfq, rfq.ReInsList) //initial RFQ email
    	if errmail != nil{
    		c.Errorf("sendemail error.")
	    }
	rfq.Modified = "modification emailed" //reinsurers emailed about modification	
	//save updated Status to datastore.
	_, err2 := datastore.Put(c, rfqKeyInfo.RfqKey, &rfq)
    if err2 != nil {
        http.Error(w, err2.Error(), http.StatusInternalServerError)
        return
    }    	
    session2, err3 := store.Get(r,"sessIns")
        if err3!=nil{
        	http.Error(w,err3.Error(),500)
        	return
        }       
    session2.Options  = &sessions.Options{
        	Path: "/",
        	MaxAge: 0, 
        	HttpOnly: true,
        }                
    //set session value        
    session2.Values["currRFQ"] = rfq
                
    session2.Save(r,w)  //note, not (w,r)
    	 
	http.Redirect(w,r,"/insdboard", http.StatusFound)	
}

