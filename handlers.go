//This defines the handler functions for basic requests.
package main

import(
	"fmt"
	"net/http"
	"html/template"
//	"errors"
	"io"
	"bytes"
	//"time"
	"appengine"
	"appengine/datastore"
	
	//"github.com/gorilla/sessions"
)


//This is the root page. The HTML appears below in entirety as guestform.
func root(w http.ResponseWriter, r *http.Request) {	
	c := appengine.NewContext(r)	
	t := template.New("index.html")
	t  = template.Must(t.ParseFiles("templates/index.html"))
	x := "nodata"
	err := t.Execute(w,x)
	if err != nil{
		c.Errorf("Error executing template")
		return		
	}		     
//io.WriteString(w, PageTop+index+PageBot)
}


//this is not being used currently. Was for testing purposes.
func view(w http.ResponseWriter, r *http.Request) {        
        io.WriteString(w, PageTop)                
        c := appengine.NewContext(r)                
        //q is a query
        q := datastore.NewQuery("RFQs").Order("RFQDate")
        b := new(bytes.Buffer)        
        //q.Run(c) runs the query q in context c
        for t := q.Run(c); ; {
        	var x RFQ
        	key, err := t.Next(&x)
        	if err == datastore.Done {
        		io.WriteString(w,"<p>Done reading</p>")
        		break
        	}
        	if err != nil {
        		//serveError(c,w,err)
        		io.WriteString(w,"<p>error is not nil</p>")
        		return
        	}
        	fmt.Fprintf(b, "<p> Key = %v\n RFQs = %#v\n\n </p>",key, x)
        	io.Copy(w,b)
        }                          
      	io.WriteString(w, PageBot)             
}

