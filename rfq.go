//entering a new request for quotes.

package main

import(
	"io"
	"net/http"
	//"net/url"
	"appengine"
	"appengine/datastore"
	//"appengine/blobstore"
	"time"
	
	"encoding/gob" //for session data
	
)

//fields starting with lower case letter won't show as entity.
//
type RFQ struct{		
	RFQId		string //10 alphanum long so 8.4*10^17 possibilities, low collision prob.
	IntID		string //optional internal ID to insurer (firm RFQ) identifier, for searchability
	Iterator	int	//just used to create numbers for table rows in dashboard
	Insurer 	string //not really necessary, but could be useful in searching/indexing/queries.
	InsEmpID	string	//employee id <email>
	InsObsList	[]string //list of insurer colleagues who can view RFQ
	ClientName	string	
	InsCat		string	//fire, group life etc
	NewCust		string `datastore:",noindex"`//renewal or new customer
	InsType		string	`datastore:",noindex"`// facultative XL, treaty XL
	Period		string	`datastore:",noindex"` //period of insurance
	EstPrem		string	`datastore:",noindex"` //estimated premium (treaty)
	Treaty		int	//binary 1 for treaty
	InsVal 		string `datastore:",noindex"`// convert to float64 //Total IDV requested by client
	InsDed		string `datastore:",noindex"`//convert to float64 Insurer's retention.
	Method		int		//comparison method. 1 is as-bid, 0 is second price 
	Brokerage	string	`datastore:",noindex"`//just for brokers
	InsRemarks  string	  `datastore:",noindex"`
	FileNames	[]string  `datastore:",noindex"`
	FileBlobKeys []appengine.BlobKey `datastore:",noindex"`//[]BlobKeys for all attached files.
	ReInsList	[]string // emails of RIs to request.
	//FollowList	[]string //emails of RIs who are followers, not price setters. 
	RFQDate		time.Time
	RFQStrTime	string `datastore:",noindex"`
	Status		int64 //1:saved, 2: submitted, 3: response rcvd (action required), 4: deal agreed, 5/6: marked deleted (shared, unshared)
	Modified	string //"new", "modified", "modification emailed"
	RespRcvdFrom []string `datastore:",noindex"` //Ids of RIs who have responded
	RespRcvdKeys []*datastore.Key `datastore:",noindex"` //the RinsRFQList key of each response	
	//LeaderResp	[]string `datastore:",noindex"`//IDs of invited leaders who have responded (treaty)
	//LeaderRespKeys []*datastore.Key `datastore:",noindex"` //keys of leaders in Treaty 
	NumRespRcvd int `datastore:",noindex"`//could be useful for column/report generation	
	InsPrivate	string `datastore:",noindex"` //Private notes to self for ins
	Median		string `datastore:",noindex"` //Median premium iff more than 2 bids.
}

//RFQ key generation
func RFQKey (c appengine.Context) *datastore.Key{
	return datastore.NewKey(c, "RFQs", "default_RFQs", 0, nil)
}

//encoding struct for session data serialization/deserialization

func init(){
	gob.Register(&RFQ{})
	gob.Register(&sessLocKey{})
	gob.Register(&RInsRFQList{})
	gob.Register(&Resp{})
}

//reads session info to get current RFQ. Type error needs to be handled better.
func SessRecallRFQ(w http.ResponseWriter, r *http.Request) (rfq RFQ, err2 error){
	session, err := store.Get(r,"sessIns")
	if err != nil{
        	http.Error(w,err.Error(),500)
        	err2 = err
        	return
         }	
	val := session.Values["currRFQ"]
	var sessrfq = &RFQ{}
	
	sessrfq, ok := val.(*RFQ)
	if !ok{
		//io.WriteString(w,PageTop)
        io.WriteString(w,"<p> Invalid session data type (sessIns).</p>")
        //io.WriteString(w,PageBot)
        return
	}
	rfq = *sessrfq
	err2 = err
return
}

type sessLocKey struct{
	RFQId	string
	RfqKey *datastore.Key
}

//session cookie to store the key of an RFQ located by insLocRFQ or RinsLocRFQ
//since these use q.GetAll, the returned keys are of type []*datastore.Key
//so when using, be sure to convert the []Key into Key first.
func SessRecLocKey(w http.ResponseWriter, r *http.Request)(key sessLocKey, err2 error){
	session, err := store.Get(r,"sessLocKey")
	if err != nil{
        	http.Error(w,err.Error(),500)
        	return
         }	
	val := session.Values["Key"]
	var sessKey = &sessLocKey{}
	
	sessKey, ok := val.(*sessLocKey)
	if !ok{
		//io.WriteString(w,PageTop)
        io.WriteString(w,"<p> Invalid session data type (sessLocKey).</p>")
        //io.WriteString(w,PageBot)
        return
	}
	key = *sessKey
	err2 = err
return
}

//struct for RI's list of RFQs for RI's dashboard and communication
//The list is based on RFQ struct with some additions and some deletions marked by //NO!
//Some fields will be stored for dashboard and querying, marked by "store"
//Others will be filled in from the relavant RFQ at run time when needed, such as to 
//display the details of an RFQ. The order has been changed a bit, to make it easier to
//observe in datastore viewer.
type RInsRFQList struct{
	RFQId		string //store
	RIUsername string	 //this is the primary Rins contact
	RinsObsList	[]string //observer list of reinsurer colleagues, can view RFQ and responses
	LeaderInv	int //indicator, 1 implies invited to leader negotiations in Treaty
	//RIUsername is out of order from RFQ struct 
	//NO! IntID		string 
	//NO! Iterator	int	
	Insurer 	string //store, it's easy.
	InsEmpID	string //store
	ClientName	string `datastore:",noindex"`
	InsCat		string //store
	NewCust		string `datastore:",noindex"`	
	InsType		string `datastore:",noindex"`
	Treaty		int	//binary 1 for treaty //treaty XL
	InsVal 		string `datastore:",noindex"`
	InsDed		string `datastore:",noindex"`
	Method		int		//1 is second price, default 0 is all bids
	EstPrem		string	`datastore:",noindex"` 
	Period		string	`datastore:",noindex"` 
	Brokerage	string	`datastore:",noindex"`//just for brokers
	InsRemarks  string	  `datastore:",noindex"`
	FileNames	[]string  `datastore:",noindex"`
	FileBlobKeys []appengine.BlobKey `datastore:",noindex"`//[]BlobKeys for all attached files.
	//NO! ReInsList	[]string  
	RFQDate		time.Time
	RFQStrTime	string `datastore:",noindex"`
	//NO! Status		int64 //1:saved, 2: submitted/in process, 3: accepted
	//NO! Modified	string //"new", "modified", "modification emailed"		
	SecondRins string	//this is the secondary who can also make modifications. Later.
	Observers	[]string //list of User's with observer status. Later.	
	ReqRcvd 	*datastore.Key  `datastore:",noindex"`//store// keys of RFQs with response required	
	RespStatus int64 //1 response needed, 2 replied, 3 deleted by insurer, 4 discarded by Rins.
	RespDate	string	`datastore:",noindex"` //date of latest response.
	RinsFileNames	[]string `datastore:",noindex"` //store //files attached by Rins
	RinsFileBlobs	[]appengine.BlobKey `datastore:",noindex"` //store  //Rins' file blobkeys
	Limit		[]string `datastore:",noindex"`//store Capacity offerred []s allow change history
	Premium		[]string `datastore:",noindex"`//store[]s allow change history
	TrtyLdrPrem float64 `datastore:",noindex"`//quote offer in leader negs
	TrtyFolQty float64 `datastore:",noindex"`//demand of follower
	Commission	[]string `datastore:",noindex"`//store []s allow change history
	Conversation []string `datastore:",noindex"`//store conversation. Combined notes so they show up as dialogue.
	ConvAuth	[]string `datastore:",noindex"` //author of conversation element.
	Private		string	`datastore:",noindex"`//store notes from Rins to self
	Median		string	`datastore:",noindex"`//iff revealed by insurer.
}

//key generation for RI's list
func RInsListKey (c appengine.Context) *datastore.Key{
	return datastore.NewKey(c, "RInsList", "default_RInsList", 0, nil)	
}

//reads session info to get current Rins version of RFQ.
func SessRecallRinsRFQ(w http.ResponseWriter, r *http.Request) (Rinsrfq RInsRFQList, err2 error){
	session, err := store.Get(r,"sessRins")
	if err != nil{
        	http.Error(w,err.Error(),500)
        	err2 = err
        	return
         }	
	val := session.Values["CurrRinsRFQ"]
	var sessrfq = &RInsRFQList{}
	
	sessrfq, ok := val.(*RInsRFQList)
	if !ok{
		//io.WriteString(w,PageTop)
        io.WriteString(w,"<p> Invalid session data type (sessRins).</p>")
        //io.WriteString(w,PageBot)
        return
	}
	Rinsrfq = *sessrfq
	err2 = err
return
}

//struct definition and template for displaying quotes to insurer in rinslocRFQ Status 3
type Resp struct {
	Rins   []string	//RIs who have responded
	RowNames []string	//rows for the table Limit, premium etc below
	Respkey	[]*datastore.Key //location of the RinsRFQList
	Limit [][]string //as many []strings as Rins, the inner ones can be whatever long
	Premium [][]string //as many []strings as Rins
	Commission [][]string
	Files [][]string
	ConvAuth [][]string
	Conversation [][]string
	Median	string
}

//reads session info to get current Response from RIs.
func SessRecResp(w http.ResponseWriter, r *http.Request) (resp Resp, err2 error){
	session, err := store.Get(r,"sessResp")
	if err != nil{
        	http.Error(w,err.Error(),500)
        	err2 = err
        	return
         }	
	val := session.Values["CurrResp"]
	var sessResp = &Resp{}
	
	sessResp, ok := val.(*Resp)
	if !ok{
		//io.WriteString(w,PageTop)
        io.WriteString(w,"<p> Invalid session data type (sessResponse).</p>")
        //io.WriteString(w,PageBot)
        return
	}
	resp = *sessResp
	err2 = err
return
}