//Tried to search the data store for given text queries.

package main

import(
	"io"
	//"fmt"
	"net/http"
	//"strings"
	"html/template"
	
	"appengine"
	"appengine/datastore"
	
	//"github.com/gorilla/mux"
)


//Function to search for arbitrary text given property and query
func SearchBase(w http.ResponseWriter, r *http.Request){
	//qry := strings.ToLower(query) //
	c := appengine.NewContext(r)
	
	r.ParseForm()
	query := r.Form.Get("query")
		
	user,err := SignInRecall(w,r)
		if err != nil{
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
	//
	io.WriteString(w, PageTop)	
	//for insurer
	if user.UserType != "Rins"{
		//results by clientname
		q1 := datastore.NewQuery("RFQs").Filter("InsEmpID =",user.Username).Filter("ClientName >=", query).Limit(10)
		var xClient []RFQ
		_, err1 := q1.GetAll(c,&xClient)
		if err1 != nil{
			http.Error(w,err.Error(),500)
		}
		
		 //set iterator for table rows.        
        for i := 0;i<len(xClient);i++{
        	xClient[i].Iterator = i+1
        }
        
		/*
		//results by category
		q2 := datastore.NewQuery("RFQs").Filter("InsEmpID =",user.Username).Filter("InsCat >=", query).Limit(1)
		var xCat []RFQ
		_, err2 := q2.GetAll(c,&xCat)
		if err2 != nil{
			http.Error(w,err.Error(),500)
		}
		
		//set iterator for table rows.        
        for i := 0;i<len(xCat);i++{
        	xCat[i].Iterator = i+1
        }
		*/
		io.WriteString(w, "<p>User <b>"+ user.Username + "</b> logged in.  <a href=\"/signout\">Logout</a></p>")
        io.WriteString(w, insSearchTop)
    	io.WriteString(w, insDashTable1)
         
        if len(xClient) == 0{
        	io.WriteString(w, "No matches found. Client names are case sensitive.")
        }
                
		t := template.New("insdash.html")
		t  = template.Must(t.ParseFiles("insdash.html"))
		//execute for each RFQ
		for _,r := range xClient{
			err3 := t.Execute(w,r)
			if err3 != nil{
				c.Errorf("Error executing template table 1")		
			}
		}
		io.WriteString(w, "</table><div>")
		
		/*
		io.WriteString(w, "<h4>Results by category</h4>")
        io.WriteString(w, "<p>Click on an RFQ ID to review, edit and send.</p>" + insDashTable1)        

		//execute for each RFQ
		for _,r := range xCat{
			err3 := t.Execute(w,r)
			if err3 != nil{
				c.Errorf("Error executing template table 1")		
			}
		}
		io.WriteString(w, "</table><div>")
		*/
		
		io.WriteString(w, PageBot)
			
	
	} else{ //for RIns
		c.Errorf("no search for type Rins")
	}

}

const insSearchTop = `

<form class="navbar-form navbar-left" action = "/search" method="post">
  
  <a class = "btn btn-primary" href="/insdboard"  role="button">View Dashboard</a> 
  <div class="form-group">
    <input type="text" name="query" class="form-control" placeholder="Search">
  </div>
  <button type="submit" class="btn btn-info">Search</button>
</form><br></br>
<h4>Results by Client Name  </h4>
<p>Click on an RFQ ID to view details.</p>
`