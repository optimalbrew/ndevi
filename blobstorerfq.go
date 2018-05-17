//Working with the blobstore for file uploads.
//Example from https://cloud.google.com/appengine/docs/go/blobstore/

package main

import(
	"html/template"
	"bytes"
	//"fmt"
	
	"time"
	"strings"
	//"strconv"
	
	"io"
	"io/ioutil"
	"net/http"
	"appengine"
	"appengine/datastore"
	"appengine/blobstore"
	"appengine/mail"
	
	"github.com/gorilla/sessions"
	"github.com/gorilla/mux"
)

func serveError(c appengine.Context, w http.ResponseWriter, err error){
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal server error.")
	c.Errorf("%v", err)
}


var rootTemplate = template.Must(template.New("root").Parse(rootTemplateHTML)) 

const rootTemplateHTML = `
<p> On most modern systems you should be able to drag and drop a file into the box. 
Please see security note below regarding permissible files types. If you upload a file by
mistake you can delete it later. </p>

<form action = "{{.}}" method = "POST" enctype = "multipart/form-data">

  	<div class="form-group">
    <label for="file">Select a file to upload:</label>
    <input type="file" name = "file[]" class = "form-control" id="file">
  	</div>
	 		
    <button type="submit" class="btn btn-primary">Upload File</button>
	</form> <br>

`    
      
const rootTemplateBot = `

<p><b>Security note:</b> Most common file formats such as PDF, Microsoft Office documents, media files (jpeg, mp3, mp4) are fine. 
For security reasons, some types of files (or zipped folders containing such formats) <i>cannot
be sent as attachments</i> to system generated emails. Some excluded types are files ending with extensions like
: <i>bat, cmd, com, exe, lib, sys </i>. A more complete list of prohibited file types is available at 
<a href = "https://cloud.google.com/appengine/docs/go/mail">Google's App Engine</a> mail services page.</p>



`

//Set the template by dynamically inserting the appropriate URL for action
func handleRoot(w http.ResponseWriter, r *http.Request){
	//check log in status
	user,err := SignInRecall(w,r)
	if err != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}	
	c := appengine.NewContext(r)
	
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, PageTop)	
	//grab the session values based on user type
	if user.UserType != "RIns"{ //insurer or broker
		rfq, err := SessRecallRFQ(w,r)	
		if err != nil{
			return
		}
	//The only reason to use rfq here is to make sure this page is shown in the context of an 
	// RFQ irrespective of user type. Else we get a upload page without proper context. 		
		io.WriteString(w,"<p>User <b>" +rfq.InsEmpID + "</b> logged in. <a href=\"/signout\">Logout</a></p>")
		io.WriteString(w, `<p>To upload files later, or when you are done adding files for now go to <a class = "btn btn-success" href="/review"  role="button">Review</a></p>`)
		
		numfiles := len(rfq.FileNames)
		io.WriteString(w, "<p><b>Files currently uploaded</b><p>")
		if numfiles == 0 {
			io.WriteString(w,"No files uploaded")
		}else{
			io.WriteString(w,"<ul>")
			for i := 0; i<numfiles;i++{
			io.WriteString(w, "<li> <a href=\"/serve/"+ rfq.FileNames[i]+ "\">"+ rfq.FileNames[i]+"</a> </li>")
			
			}
			io.WriteString(w, "</ul>")
		}
		
		uploadURL, err := blobstore.UploadURL(c, "/upload",nil)
		if err != nil{
			serveError(c, w, err)
			return
		}
		err = rootTemplate.Execute(w,uploadURL)
		if err != nil{
			c.Errorf("%v", err)
			return
		}
		io.WriteString(w, ` <p>Done adding files? <a class = "btn btn-success" href="/review"  role="button">Review</a></p>`)		
	}else{		//for reinsurers still called 'rfq'
		rfq, err := SessRecallRinsRFQ(w,r)	
		if err != nil{
				return
			}									
		io.WriteString(w,"<p>User <b>" +rfq.RIUsername + "</b> logged in. <a href=\"/signout\">Logout</a></p>")
		
		numfiles := len(rfq.RinsFileNames)
		io.WriteString(w, "<p><b>Files currently uploaded</b><p>")
		if numfiles == 0 {
			io.WriteString(w,"No files uploaded")
		}else{
			io.WriteString(w,"<ul>")
			for i := 0; i<numfiles;i++{
			io.WriteString(w, "<li> <a href=\"/serverins/"+ rfq.RIUsername+"/"+ rfq.RinsFileNames[i]+ "\">"+ rfq.RinsFileNames[i]+"</a> </li>")
			
			}
			io.WriteString(w, "</ul>")
		}
		
		
		uploadURL, err := blobstore.UploadURL(c, "/uploadrins",nil)
		if err != nil{
			serveError(c, w, err)
			return
		}
		err = rootTemplate.Execute(w,uploadURL)
		if err != nil{
			c.Errorf("%v", err)
			return
		}
		io.WriteString(w, ` <p>Done adding files? <a class = "btn btn-success" href="/reviewrins"  role="button">Review</a></p>`)	
	}
	io.WriteString(w, rootTemplateBot + PageBot)		
}	


//Handle the upload to blobstore, post processing. For insurers. 
func handleUpload(w http.ResponseWriter, r *http.Request){
	//check log in status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//recall the current RFQ from session cookie
	rfq, err := SessRecallRFQ(w,r)	
	if err != nil{
		return
	}
	//recall the rfq's location in data store, so won't have to query for it.
	//The id is still .RFQId, and the key is .RfqKey
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
			
	c := appengine.NewContext(r)
	//blobs is a map["form element"][]*BlobInfo, form element is "file", of course, file[] used in html code.
	blobs, _, err := blobstore.ParseUpload(r) // blobs is map[string][]*Blobinfo	
	if err != nil {
		serveError(c, w, err)
		return
	}
	//this is the mapped value for file, there should be as many entries in the map as files uploaded
	//each of which is a BlobInfo struct type. The keys are accessed through ".BlobKey"
	file := blobs["file[]"] //file is a map
		//numfiles := len(file) 	
	if len(file) == 0{  //len(file) is the number of files uploaded
		c.Errorf("no files uploaded.") //this shows up on console.
		//http.Redirect(w, r, "/handleroot", http.StatusFound)
		return	
	} else{
	//Add the blob keys to the relevant RFQ, which may have 0 or more files currently.			
			currNumFiles := len(rfq.FileNames)//current number of files
			numFilesToAdd := len(file)
			tempBlobKeys := make([]appengine.BlobKey,currNumFiles + numFilesToAdd)
			tempFilesNames := make([]string, currNumFiles + numFilesToAdd)
		
			//copy existing info into temp slices
			for i := 0;i<currNumFiles;i++{
				tempBlobKeys[i] = rfq.FileBlobKeys[i]
				tempFilesNames[i] = rfq.FileNames[i]
			}
			//Now copy in the new files' info 		
			j := 0
			for i := 0 ; i<numFilesToAdd;i++{				
					if string(file[i].Filename)!= ""{
						tempBlobKeys[currNumFiles+j] = file[i].BlobKey
						tempFilesNames[currNumFiles+j] = string(file[i].Filename)
						j++
					}				
			}	
			//update the rfq
			rfq.FileBlobKeys = tempBlobKeys
			rfq.FileNames = tempFilesNames								
			
			if rfq.Status == 2{ 
				rfq.Modified = "modified"
			}
			
			//Update RFQ session cookie
			session, err := store.Get(r,"sessIns")
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
        	session.Values["currRFQ"] = rfq 
	     	session.Save(r,w)  //note, not (w,r)
			
			//Now save the updated information to datastore using the key from cookie.
			//rfqKeyInfo.RfqKey 
			_, errPut := datastore.Put(c, rfqKeyInfo.RfqKey, &rfq)
    		if errPut != nil {
            	http.Error(w, errPut.Error(), http.StatusInternalServerError)
            	return
    		}

		//reload page to allow further file uploads.
		http.Redirect(w,r,"/handleroot",http.StatusFound)
				
	}
	return
}

//handle uploads for Rins. Pretty much the same as that for ins, just separated to avoid  mistakes.
//Handle the upload to blobstore, post processing. For insurers. 
//Todo: Some intimation to insurers about file loads, if response already sent. Most likely
//conversation will take care of it.
func handleUploadRins(w http.ResponseWriter, r *http.Request){
	//check log in status
	_,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/",http.StatusFound)
			return
		}
	//recall the current RFQ from session cookie
	rfq, err := SessRecallRinsRFQ(w,r)	
	if err != nil{
		return
	}
	//recall the rfq's location in data store, so won't have to query for it.
	//The id is still .RFQId, and the key is .RfqKey
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}
			
	c := appengine.NewContext(r)
	//blobs is a map["form element"][]*BlobInfo, form element is "file", of course, file[] used in html code.
	blobs, _, err := blobstore.ParseUpload(r) // blobs is map[string][]*Blobinfo	
	if err != nil {
		serveError(c, w, err)
		return
	}
	//this is the mapped value for file, there should be as many entries in the map as files uploaded
	//each of which is a BlobInfo struct type. The keys are accessed through ".BlobKey"
	file := blobs["file[]"] //file is a map
		//numfiles := len(file) 	
	if len(file) == 0{  //len(file) is the number of files uploaded
		c.Errorf("no files uploaded.") //this shows up on console.
		//http.Redirect(w, r, "/handleroot", http.StatusFound)
		return	
	} else{
	//Add the blob keys to the relevant RFQ, which may have 0 or more files currently.			
			currNumFiles := len(rfq.RinsFileNames)//current number of files
			numFilesToAdd := len(file)
			tempBlobKeys := make([]appengine.BlobKey,currNumFiles + numFilesToAdd)
			tempFilesNames := make([]string, currNumFiles + numFilesToAdd)
		
			//copy existing info into temp slices
			for i := 0;i<currNumFiles;i++{
				tempBlobKeys[i] = rfq.RinsFileBlobs[i]
				tempFilesNames[i] = rfq.RinsFileNames[i]
			}
			//Now copy in the new files' info 		
			j := 0
			for i := 0 ; i<numFilesToAdd;i++{				
					if string(file[i].Filename)!= ""{
						tempBlobKeys[currNumFiles+j] = file[i].BlobKey
						tempFilesNames[currNumFiles+j] = string(file[i].Filename)
						j++
					}				
			}	
			//update the rfq (session version)
			rfq.RinsFileBlobs = tempBlobKeys
			rfq.RinsFileNames = tempFilesNames								
			
			
			//update RFQ from RinsRFQList: Have to grab it first.
			var realRinsRFQ RInsRFQList
			errReal := datastore.Get(c,rfqKeyInfo.RfqKey, &realRinsRFQ)
			if errReal != nil{
				c.Errorf("could not grab real RFQ from RinsRFQLIst")
				return	
			}
			realRinsRFQ.RinsFileBlobs = tempBlobKeys
			realRinsRFQ.RinsFileNames = tempFilesNames
			//Now save the updated information to datastore using the key from cookie.
			//rfqKeyInfo.RfqKey 
			_, errPut := datastore.Put(c, rfqKeyInfo.RfqKey, &realRinsRFQ)
    		if errPut != nil {
            	http.Error(w, errPut.Error(), http.StatusInternalServerError)
            	return
    		}
			//if rfq.Status == 2{ 
			//	rfq.Modified = "modified"
			//}
			
			//Update Rins' version of RFQ session cookie
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
    		session.Values["CurrRinsRFQ"] = rfq 
	     	session.Save(r,w)  //note, not (w,r)
			
			
			//reload page to allow further file uploads.
			http.Redirect(w,r,"/handleroot",http.StatusFound)
		
	}
	return
}



//modified from Google example. //Serving blobs depending on context, uses session data for now
//this version is to handle links to files uploaded by insurers.
func handleServe(w http.ResponseWriter, r *http.Request){
	//place access control here. 
	usr, err := SignInRecall(w,r)	
	if err != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
	//session data may not be of type RFQ, it may be of type RinsListRFQ, change accordingly	
	if usr.UserType != "RIns"{ //insurer or broker
		rfq, err2 := SessRecallRFQ(w,r)	
		if err2 != nil{
			return
		}	
		str2blob := make(map[string]appengine.BlobKey)
		for i:=0;i<len(rfq.FileNames);i++{
			str2blob[rfq.FileNames[i]] = rfq.FileBlobKeys[i]
		}
		vars := mux.Vars(r)
		blobKey := str2blob[vars["key"]] //key could be blobkey itself or a temp string to represent it
		//func Send(response http.ResponseWriter, blobKey appengine.BlobKey)
 		blobstore.Send(w, blobKey) 								
	}else{ //if it is reinsurer
		rfq, err2 := SessRecallRinsRFQ(w,r)	
		if err2 != nil{
			return
		}
		str2blob := make(map[string]appengine.BlobKey)
		for i:=0;i<len(rfq.FileNames);i++{
			str2blob[rfq.FileNames[i]] = rfq.FileBlobKeys[i]
		}			
		vars := mux.Vars(r)
		blobKey := str2blob[vars["key"]] //key could be blobkey itself or a temp string to represent it
		//func Send(response http.ResponseWriter, blobKey appengine.BlobKey)
 		blobstore.Send(w, blobKey) 		
	}
return
}

//this version is to handle links to files uploaded by REinsurers.
//this is more complex, as there will be files uploaded by differes REs for the same RFQ.
//having a different serve function also reduces or eliminates the risk of revealing a RE's
//docs to other REs.
func handleServeRins(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	//place access control here. 
	usr, err := SignInRecall(w,r)	
	if err != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
	
	if usr.UserType != "RIns"{ //insurer or broker
		// this is more complex, there will be multiple reinsurers, this implies
		//using two r.mux keys: {RinsId} and {FileId}		
		//first recall the insurers' RFQ from cookie
		rfq, err2 := SessRecallRFQ(w,r)	
		if err2 != nil{
			return
		}
		//then find which reinsurers are listed and map their names to their data RinsRFQ
		//get that RinsRFQ, then use a map of file names and blobkeys
		rins2rinsRFQList := make(map[string]*datastore.Key)
		//if Ins is viewing a response, then must know the location of RinsRFQ. So it must be
		//available in its own RFQ struct as response received properties (responders and the
		//location of their response in RinsRFQList).
		for i:= 0; i<len(rfq.RespRcvdFrom);i++{
			rins2rinsRFQList[rfq.RespRcvdFrom[i]] = rfq.RespRcvdKeys[i] 
		}
		vars := mux.Vars(r)
		RinsListKey := rins2rinsRFQList[vars["rinsId"]]
		//grab the relevant RE's rfq from RinsRFQList
		var Rinsrfq RInsRFQList
		errGet := datastore.Get(c, RinsListKey, &Rinsrfq)
		if errGet != nil {
			c.Errorf("Get failed for RinsRFQList in serveRins")
			return
		}
		//if it works create the map for the RE's files and associated keys	
		str2blob := make(map[string]appengine.BlobKey)
		for i:=0;i<len(Rinsrfq.RinsFileNames);i++{
			str2blob[Rinsrfq.RinsFileNames[i]] = Rinsrfq.RinsFileBlobs[i]
		}
		blobKey := str2blob[vars["fileId"]] //key could be blobkey itself or a temp string to represent it
		//func Send(response http.ResponseWriter, blobKey appengine.BlobKey)
 		blobstore.Send(w, blobKey) 								
	}else{ //if it is reinsurer: This is straightforward
		vars := mux.Vars(r)		
		//check that this is the same reinsurer as the one logged in
		if usr.Username != vars["rinsId"]{
			c.Errorf("unauthorized user in serveRins")
			return
		}
		Rinsrfq, err2 := SessRecallRinsRFQ(w,r)	
		if err2 != nil{
			return
		}
		str2blob := make(map[string]appengine.BlobKey)
		for i:=0;i<len(Rinsrfq.RinsFileNames);i++{
			str2blob[Rinsrfq.RinsFileNames[i]] = Rinsrfq.RinsFileBlobs[i]
		}			
		blobKey := str2blob[vars["fileId"]] //key could be blobkey itself or a temp string to represent it
		//func Send(response http.ResponseWriter, blobKey appengine.BlobKey)
 		blobstore.Send(w, blobKey) 		
	}
return
}



//function to remove uploaded files by insurer
//Files are removed based on file name, if multiple ones have the same name, 
//then the map is not well defined, the first matching name gets deleted. 
func RemoveFileIns(w http.ResponseWriter, r *http.Request){
	//place access control here. 
	c := appengine.NewContext(r)
	usr, err := SignInRecall(w,r)	
	if err != nil{	
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
		
	if usr.UserType == "RIns"{ //don't allow rins to remove
		c.Errorf("Rins cannot delete Ins file")
		io.WriteString(w, PageTop)
		io.WriteString(w, "You are not authorized to remove this file.")
		io.WriteString(w, PageBot)
		return
	}
	
	//check to verify this is the most recent login	
	errStale := StaleLog(c,usr)
	if errStale != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		c.Errorf("stale login")
		return
	}
	
	//session data may not be called rfq, change accordingly
	rfq, err2 := SessRecallRFQ(w,r)	
	if err2 != nil{
		return
	}		
	
	//rfq key cookie rfqKeyInfo.RfqKey So RFQ can be updated with deleted information
	rfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}	
	numFiles := len(rfq.FileNames)
	//pick the filename from URL
	vars := mux.Vars(r)
	fileToDel := vars["key"]
	//define map from filename to blobkey
	str2blob := make(map[string]appengine.BlobKey)
	//fill the map
	for i:=0;i<numFiles;i++{
		str2blob[rfq.FileNames[i]] = rfq.FileBlobKeys[i]
	}
	//use map to identify key
	blobKey := str2blob[fileToDel] //blobkey to delete blob
	
	//delete blob
	errblob := blobstore.Delete(c, blobKey)
	 	if errblob != nil{
     		c.Errorf("Could not delete blobfile: %v", errblob)
     		return
     	}	
		
	//Find the location of the file in RFQ
	var j int
	for i:=0;i<numFiles;i++{
		if rfq.FileNames[i] == fileToDel{
			j = i
			break
		}			
	}
	//update rfq to reflect change
	tempfileList := make([]string,numFiles-1) //slices are one less than numfiles
	tempfileBlobList := make([]appengine.BlobKey,numFiles-1)

	for i:=0; i<j;i++{
		tempfileList[i] = rfq.FileNames[i]
		tempfileBlobList[i] = rfq.FileBlobKeys[i]
	}
	for i:=j+1; i<numFiles;i++{
		tempfileList[i-1] = rfq.FileNames[i]
		tempfileBlobList[i-1] = rfq.FileBlobKeys[i]
	}
	rfq.FileNames = tempfileList
	rfq.FileBlobKeys = tempfileBlobList
	
	//if rfq.Status == 2{ 	//may be puzzling for reinsurers, if there is no comment
	//	rfq.Modified = "modified"
	//}
	
	//save updated Status to datastore.
	_, errPut := datastore.Put(c, rfqKeyInfo.RfqKey, &rfq)
    if errPut != nil {
        http.Error(w, errPut.Error(), http.StatusInternalServerError)
        return
    } 
    //update session cookie for RFQ
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

	//For status 2 do something similar to share (but be careful about associations)	
	//send to relevant review page.
		if rfq.Status == 1 || rfq.Status == 2 {  //no responses
			t := template.New("insMainRFQ.html")
			t  = template.Must(t.ParseFiles("templates/insMainRFQ.html"))
			err = t.Execute(w,rfq)
			if err != nil{
				c.Errorf("Error executing template")
			return		
			}
		} else if rfq.Status == 3 { //already shared with insurers. Deleted doesn't matter.
			
			resp, errResp := SessRecResp(w,r) //used *resp instead of just resp
			if errResp != nil{
				c.Errorf("could not read sessResp cookie in delfileIns")
				return
			}	
			
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
return
}

//function to remove uploaded files by Reinsurer
//Files are removed based on file name, if multiple ones have the same name, 
//then the map is not well defined, the first matching name gets deleted. 
func RemoveFileRins(w http.ResponseWriter, r *http.Request){ 
	c := appengine.NewContext(r)
	usr, err := SignInRecall(w,r)	
	if err != nil{	
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}		
	if usr.UserType != "RIns"{ //don't allow rins to remove
		c.Errorf("Only Rins can delete this file")
		io.WriteString(w, PageTop)
		io.WriteString(w, "You are not authorized to remove this file.")
		io.WriteString(w, PageBot)
		return
	}	
	//check to verify this is the most recent login	
	errStale := StaleLog(c,usr)
	if errStale != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		c.Errorf("stale login")
		return
	}
	
	//session data may not be called rfq, change accordingly
	rfq, err2 := SessRecallRinsRFQ(w,r)	
	if err2 != nil{
		return
	}			
	//rfq key cookie rfqKeyInfo.RfqKey So RFQ can be updated with deleted information
	RinsrfqKeyInfo, errKeyInfo := SessRecLocKey(w,r)	
	if errKeyInfo != nil{
		return
	}	
	numFiles := len(rfq.RinsFileNames)
	//pick the filename from URL
	vars := mux.Vars(r)
	fileToDel := vars["key"]
	//define map from filename to blobkey
	str2blob := make(map[string]appengine.BlobKey)
	//fill the map
	for i:=0;i<numFiles;i++{
		str2blob[rfq.RinsFileNames[i]] = rfq.RinsFileBlobs[i]
	}
	//use map to identify key
	blobKey := str2blob[fileToDel] //blobkey to delete blob
	
	//delete blob
	errblob := blobstore.Delete(c, blobKey)
	 	if errblob != nil{
     		c.Errorf("Could not delete blobfile in RemoveFileRins: %v", errblob)
     		return
     	}	
		
	//Find the location of the file in RFQ
	var j int
	for i:=0;i<numFiles;i++{
		if rfq.RinsFileNames[i] == fileToDel{ //BUG: multiple files have the same name, could delete the wrong one.
			j = i
			break
		}			
	}
	//update rfq to reflect change
	tempfileList := make([]string,numFiles-1) //slices are one less than numfiles
	tempfileBlobList := make([]appengine.BlobKey,numFiles-1)

	for i:=0; i<j;i++{
		tempfileList[i] = rfq.RinsFileNames[i]
		tempfileBlobList[i] = rfq.RinsFileBlobs[i]
	}
	for i:=j+1; i<numFiles;i++{
		tempfileList[i-1] = rfq.RinsFileNames[i]
		tempfileBlobList[i-1] = rfq.RinsFileBlobs[i]
	}
	rfq.RinsFileNames = tempfileList
	rfq.RinsFileBlobs = tempfileBlobList
	
	
	//save updated Status to datastore.
	//Grab a copy first.
	var realRinsrfq RInsRFQList
	errGet := datastore.Get(c, RinsrfqKeyInfo.RfqKey, &realRinsrfq)
	if errGet != nil{
		c.Errorf("couldn't Get Rins RFQ from RinsRFQList in RemoveFileRins")
		return
	}
	//make sure they match
	if realRinsrfq.RFQId != rfq.RFQId || realRinsrfq.RIUsername != rfq.RIUsername{
		c.Errorf("wrong RFQ in RemoveFileRins")
		return	
	}		
	realRinsrfq.RinsFileNames = rfq.RinsFileNames
	realRinsrfq.RinsFileBlobs = rfq.RinsFileBlobs
		
	_, errPut := datastore.Put(c, RinsrfqKeyInfo.RfqKey, &realRinsrfq)
	if errPut != nil{
		c.Errorf("couldn't put updated Rins RFQ into RinsRFQList in RemoveFileRins")
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
    session.Values["CurrRinsRFQ"] = rfq
    session.Save(r,w)
	
	//then go back to the view
    t := template.New("RinsMainRFQ.html")
	t  = template.Must(t.ParseFiles("templates/RinsMainRFQ.html"))
	err = t.Execute(w,rfq)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}
return
}


//function to mark an RFQ as closed. Just a status change to RFQ.Status = 5 (shared), 6 (unshared) and
//RinsRFQList.RespStatus = 3
func MarkRFQdel(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	usr, err1 := SignInRecall(w,r)	 //must be signed in
	if err1 != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
	if usr.UserType == "RIns"{ //must not be reinsurer, cannot delete RFQ (could disregard it)
		c.Errorf("Rins cannot delete rfq")
		return
	}
	rfq, err2 := SessRecallRFQ(w,r)	
	if err2 != nil{
		return
	}
	locKeyStruct, err3 := SessRecLocKey(w,r)	
	if err3 != nil{
		return
	}
	locKey := locKeyStruct.RfqKey
		
	var rinsFindrfq []RInsRFQList
	q := datastore.NewQuery("RInsList").Filter("InsEmpID =", usr.Username).Filter("RFQId =", rfq.RFQId)       //Filter("ReqRcvd =",locKey).    
	//Locate RinsList Keys either filter should do, this is just additional security.
	RinsListKeys, err4 := q.GetAll(c, &rinsFindrfq)
	if err4 != nil{
		c.Errorf("RFQ not found in RInsList in MarkRFQdel")
		return
	}	        
	if len(RinsListKeys)>0{ //if it has already been shared				
		//set the status to deleted
		for i := 0;i<len(RinsListKeys);i++{
			rinsFindrfq[i].RespStatus  = 3 //mark as deleted by insurer
			rinsFindrfq[i].RFQDate = time.Now() //reference for deletion post a certain #days.			
			_, errPut := datastore.Put(c, RinsListKeys[i], &rinsFindrfq[i])
			if errPut != nil{
				c.Errorf("Could not update RFQ deleted status to Rinslist entities in MarkRFQdel")
			}
		}
    }
    //Update own RFQ status to 5 or 6 as the case may be.
    if rfq.Status == 2 || rfq.Status == 3 { //shared
    	rfq.Status = 5
    }else if rfq.Status == 1{ //never shared
    	rfq.Status = 6
    }
    rfq.RFQDate = time.Now() //update the time (reference for deletion post a certain #days)
    rfq.RFQStrTime = rfq.RFQDate.Format("2 Jan 2006 15:04 UTC")
    rfq.NumRespRcvd = len(rfq.RespRcvdFrom)
    //upload it to RFQ
    _, err6 := datastore.Put(c, locKey, &rfq)
    if err6 != nil {
    	c.Errorf("could not put rfq into RFQ in MarkRFQdel")
    	return
    }    
	http.Redirect(w,r,"/insdboard", http.StatusFound)
}


//function to recover an RFQ marked as deleted. System adds note about reactivation and
//places it in respstatus 1 (Rins's response required).
func RecoverRFQdel(w http.ResponseWriter, r *http.Request){
c := appengine.NewContext(r)
	usr, err1 := SignInRecall(w,r)	 //must be signed in
	if err1 != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
	if usr.UserType == "RIns"{ //must not be reinsurer, cannot delete RFQ (could disregard it)
		c.Errorf("Rins cannot recover rfq")
		return
	}
	rfq, err2 := SessRecallRFQ(w,r)	
	if err2 != nil{
		return
	}
	locKeyStruct, err3 := SessRecLocKey(w,r)	
	if err3 != nil{
		return
	}
	locKey := locKeyStruct.RfqKey
	//delete RFQ from RInsList, or rather, mark it as deleted. Leave it to Rins too delete.
	var rinsFindrfq []RInsRFQList
	if rfq.Status == 5{ //deleted a shared RFQ earlier		
		q := datastore.NewQuery("RInsList").Filter("InsEmpID =", usr.Username).Filter("RFQId =", rfq.RFQId)           
		//Locate RinsList Keys either filter should do, this is just additional security.
			RinsListKeys, err4 := q.GetAll(c, &rinsFindrfq)
			if err4 != nil{
				c.Errorf("RFQ not found in RInsList in RecoverRFQdel")
				return
			}	        
		if len(rinsFindrfq)>0{
			//set the status to recovered and response required and add a message.
			auth := rfq.InsEmpID
			msg := "[System generated messge]: Closed RFQ reactivated by insurer." 
		
			for i := 0;i<len(RinsListKeys);i++{
				rinsFindrfq[i].RespStatus  = 1 //mark as recovered by insurer
				rinsFindrfq[i].ConvAuth = strArrAppZero(rinsFindrfq[i].ConvAuth, auth)
				convArr := []string{msg, " ", "(",time.Now().Format("2 Jan 2006 15:04 UTC"),")"}
				convStr := strings.Join(convArr, "")	
				rinsFindrfq[i].Conversation = strArrAppZero(rinsFindrfq[i].Conversation, convStr)
				_, errPut := datastore.Put(c, RinsListKeys[i], &rinsFindrfq[i])
				if errPut != nil{
					c.Errorf("Could not update RFQ recovered status to Rinslist entities in RecoverRFQdel")
				}
			}
		 }
	}    
    
    if rfq.Status == 5{  //had deleted a shared rfq 
    	rfq.InsRemarks = "[System generated]: Closed RFQ reactivated by insurer." 
    	if len(rfq.RespRcvdFrom) == 0{
	    	rfq.Status = 2
	    }else{
		    rfq.Status = 3
	    }
	}else if rfq.Status == 6{ //had deleted unshared
		rfq.Status = 1
	}
	
    _, err6 := datastore.Put(c, locKey, &rfq)
    if err6 != nil {
    	c.Errorf("could not put rfq into RFQ in MarkRFQdel")
    	return
    }    
	http.Redirect(w,r,"/insdboard", http.StatusFound)
}





//Function to delete all RFQ's (and all associated files) including RI files
//marked for deletion and that meet the time requirement.
func DeleteBatchRFQ(w http.ResponseWriter, r *http.Request) (errBatch error){
	c := appengine.NewContext(r)
	
	//CANNOT use cookies yet, since they have not been set. This is at the end of signin.
	//At present anyway. 
	usr, err1 := SignInRecall(w,r)	 //must be signed in
	if err1 != nil{
		http.Redirect(w,r,"/",http.StatusFound)
		return
	}
	if usr.UserType == "RIns"{ //must not be reinsurer, cannot delete RFQ (could disregard it)
		//c.Errorf("Rins cannot delete rfq")
		return
	}
	//find all the RFQs for this user that are marked for deletion, that is rfq.Status 5/6	
	var rfqDel5, rfqDel6 []RFQ
	qDel5 := datastore.NewQuery("RFQs").Filter("InsEmpID =", usr.Username).
										  Filter("Status =", 5).
										  Filter("RFQDate <", time.Now().Add(-OneWeek))	
										 //Filter("RFQDate <", time.Now().Add(-OneWeek)) 	
	keysDel5, errDel5 := qDel5.KeysOnly().GetAll(c,&rfqDel5)
	if errDel5 != nil{
		errBatch = errDel5
		//c.Errorf("could not find RFQs with stat 5 to delete in batch mode")
	}	
	
	qDel6 := datastore.NewQuery("RFQs").Filter("InsEmpID =", usr.Username).
										  Filter("Status =", 6).
										  Filter("RFQDate <", time.Now().Add(-OneWeek))
										  //Filter("RFQDate <", time.Now().Add(-OneWeek))	
	
	keysDel6, errDel6 := qDel6.KeysOnly().GetAll(c,&rfqDel6)
	if errDel6 != nil{
		errBatch = errDel5
		//c.Errorf("could not find RFQs with stat 6 to delete in batch mode")
	}
	
	if len(keysDel5)+len(keysDel6) == 0 {
		//c.Errorf("Nothing to delete yet in batchRFQdel")
		return
	}
	keysToDel := make([]*datastore.Key, len(keysDel5)+len(keysDel6))	
	
	if len(keysDel5) !=0 { 
		for i := 0; i<len(keysDel5);i++{
			keysToDel[i] = keysDel5[i]
		}
	}
	if len(keysDel6) != 0{
		for i := 0; i<len(keysDel6);i++{
			keysToDel[len(keysDel5)+i] = keysDel6[i]
		} 
	} 
	
	//Find and delete all associated RinsRFQs, and all associated files
    for i := 0; i<len(keysToDel);i++{
    	//grab the RFQ to be deleted. It has to be loaded to delete its blobs anyway.
    	var y RFQ
		errGet := datastore.Get(c, keysToDel[i] ,&y)
		if errGet != nil {
			errBatch = errGet
			return
			//c.Errorf("could not get RFQ from store in batchRFQdel")
		}    	
    	//Query RinsRFQList: overkill, but match InsID, Key, RespStatus	
    	var x []RInsRFQList
    	qRinsList := datastore.NewQuery("RInsList").Filter("InsEmpID =", usr.Username).Filter("RFQId =", y.RFQId) //
    	//could also use response status and RFQ key as filters. Tried them earlier (didn't work).
    	keyRinsList, errRinsList := qRinsList.GetAll(c, &x)
    	if errRinsList != nil{    	
    		errBatch = errRinsList
    		c.Errorf("could not find any Rinslist RFQ to delete for index i above in batchRFQdel")
    	}
    	
    	if len(keyRinsList)>0 { //some Rins are associated with this request
			for j := 0; j<len(keyRinsList);j++{ //for each identified RFQ in rins list
				//delete all files blobs associated with each found match
				if len(x[j].RinsFileBlobs) > 0{
					errBlobDel := blobstore.DeleteMulti(c, x[j].RinsFileBlobs)
					if errBlobDel != nil{
						errBatch = errBlobDel
						c.Errorf("Blob deletion failed in Rins for batchRFQdel")
						return //should not proceed, would leave blobs orphaned
					}
				}			
			}    	
    		//delete the matched entities themselves from RinsList
    		errRinsListDel := datastore.DeleteMulti(c, keyRinsList)
    		if errRinsListDel != nil{
    			errBatch = errRinsListDel
    			c.Errorf("could not delete RinsList entities in batchRFQdel")
    		}	
    	}   
		//To delete the files associated with the RFQ itself
		//delete any associated ins files
		if len(y.FileBlobKeys)>0 { 
			errInsBlob := blobstore.DeleteMulti(c, y.FileBlobKeys)
			if errInsBlob != nil{
				errBatch = errInsBlob 
				c.Errorf("could not delete ins RFQ blobs")
				return //should not delete RFQ , would leave blobs orphaned.
			}
		}
		//delete the entity itself
		errDelRfq := datastore.Delete(c, keysToDel[i])
		if 	errDelRfq != nil{
			errBatch = errDelRfq 
			c.Errorf("could not delete ins RFQ entity")
		}
    } 
return
}



//blobstore functions to remove selected uploaded files
// blobstore.Delete(c context.Context, blobKey appengine.BlobKey) error
// blobstore.DeleteMulti(c context.Context, blobKey []appengine.BlobKey) error


//take c and rfq as given and email to RIs
//third argument is NOT redundant. It is useful in situations like adding
//Rins to existing RFQ, when it does not make sense to send a new email
//to entire ReInsList in current RFQ
func sendEmailRFQ(c appengine.Context, rfq RFQ, RinsEmailList []string) (error) {
		
		numfiles := len(rfq.FileNames)
		
		attach :=  make([]mail.Attachment,numfiles)
		//create the attachment struct
		for i := 0; i<numfiles; i++ {
			blobKey := rfq.FileBlobKeys[i]			
			blobInfo, err := blobstore.Stat(c, blobKey)
			if err != nil{
				c.Errorf("Could not load blobinfo")
				break
			} 
			reader := blobstore.NewReader(c, blobKey)
			val, err2 := ioutil.ReadAll(reader)
			if err2 != nil	{
			c.Errorf("Could not read file")
			break
			//return
			}
			attach[i] = mail.Attachment{
					Name: blobInfo.Filename , //string
        			Data: val, //[]byte
        			ContentID: blobInfo.ContentType , //string
        			}		        			
		}				
		//create the message and send email		
			var err4 error	
			
			var sub string
			sub ="Reinsurance RFQ (Id: "+rfq.RFQId+" via ndevi)"
				//bcc version
	
const bodyTemp = `This is a system generated {{if .Modified}} {{.Modified}} {{end}} reinsurance quote request sent on behalf of <{{.InsEmpID}}>. Please direct all replies to <{{.InsEmpID}}>, and not to us as replies are not monitored.
			
Summary: The request is for "{{.InsCat}}" and the type is "{{.InsType}}." The proposed insured value is INR {{.InsVal}} and the insurer's retention is {{.InsDed}}. 
			
{{if .InsRemarks}}
The insurer representative has added the following remarks: 
"{{.InsRemarks}}" {{/* Do not use {{.}} here, else it prints the entire struct! */}}			
{{end}}	
		
We encourage you to register at www.ndevi.com, where you can view all requests sent to your email ID from various insurers in a single place, respond to them, observe latest modifications, and interact with insurers over a secure and convenient web interface.

We aim to minimize the number of emails people receive from us. Our default policy is to send automatic emails (with attachments) to reinsurers only for new RFQs. Emails regarding modifications are sent only when specifically requested by insurers.

In the near future, we hope to send only infrequent "digests" that summarize the status of RFQs received and responded to. We will also work with reinsurers who choose to opt out from receiving emails from us.` 
			
			t := template.Must(template.New("bodyTemp").Parse(bodyTemp))
			var doc bytes.Buffer 
        	errTem := t.Execute(&doc,rfq)
        	if errTem !=nil{
        		c.Errorf("%v",errTem)
        	} 
        	body := doc.String() 
			 
				
				//msg struct
				msg := &mail.Message{
        		Sender: //deleted email
       			ReplyTo: rfq.InsEmpID,
        		To: []string{rfq.InsEmpID},
        		Bcc: RinsEmailList,// and NOT rfq.ReInsList, sometimes we only want to email subset
        		Subject: sub,
        		Body: body,   
        		Attachments: attach[0:],//[]mail.Attachment{attach},                        
        		}     
        		
        		if err4 = mail.Send(c,msg); err4 != nil {
        		c.Errorf("Email not sent")	
        		}
			
			
			//individual email version commented out (this will result in multiple copies sent to insurer.
			/*
			for i:=0;i<len(RinsEmailList);i++{
				rcpt := []string{RinsEmailList[i]}  //email rcpt list
		        
    	    	msg := &mail.Message{
        		Sender: //deleted email
       			ReplyTo: rfq.InsEmpID,
        		To: rcpt,
        		//Bcc: RinsEmailList, //not from rfq, sometimes we may wish to email subset
        		Subject: "Reinsurance quote request",
        		Body: "Here is a request for.",   
        		Attachments: attach[0:],//[]mail.Attachment{attach},                        
        		}     
        		
        		if err4 = mail.Send(c,msg); err4 != nil {
        		c.Errorf("Email not sent")	
        		}
        	}               			
			*/
	return err4
}

//email to inform of modification in RFQ
func modEMailRFQ(c appengine.Context, rfq RFQ, RinsEmailList []string) (error) {
		
		//create the message and send email		
			var err4 error	
			
			var sub string
			sub ="Reinsurance RFQ modified (Id: "+rfq.RFQId+" via ndevi)"
				//bcc version
	
const bodyTempMod = `This is a system generated email sent on behalf of <{{.InsEmpID}}>. Please direct all replies to <{{.InsEmpID}}>, and not to us as replies are not monitored.

There has been a modification to an existing reinsurance quote request (ndevi.com id {{.RFQId}}).
			
{{if .InsRemarks}}
The insurer representative has added the following remarks concerning the change: 
"{{.InsRemarks}}" {{/* not {{.}} here, else it prints the entire struct! */}}			
{{end}}			
We encourage you to register at www.ndevi.com, where you can view all requests sent to your email ID from various insurers in a single place, and respond to all them over a secure, convenient web interface.` 
			
			t := template.Must(template.New("bodyTempMod").Parse(bodyTempMod))
			var doc bytes.Buffer 
        	errTem := t.Execute(&doc,rfq)
        	if errTem !=nil{
        		c.Errorf("%v",errTem)
        	} 
        	body := doc.String() 
			 
				
				//msg struct
				msg := &mail.Message{
        		Sender: // deleted email
       			ReplyTo: rfq.InsEmpID,
        		To: []string{rfq.InsEmpID},
        		Bcc: RinsEmailList,// and NOT rfq.ReInsList, sometimes we only want to email subset
        		Subject: sub,
        		Body: body,   
        		//Attachments: attach[0:],//[]mail.Attachment{attach},                        
        		}     
        		
        		if err4 = mail.Send(c,msg); err4 != nil {
        		c.Errorf("Email not sent")	
        		}
	return err4
}

