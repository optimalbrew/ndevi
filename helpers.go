//some funcs to assist in the code: Random Strings, Time Since (days/Hours)

package main

import(
	"math/rand"
	"time"
	"strconv"
	"sort"
	"log"
	"appengine/datastore"
	"errors"
)
	
var medianerror error = errors.New("Median not reported when number of bids is less than 3.")

func init(){
	rand.Seed(time.Now().UnixNano())
}

//for use in deleting RFQs and associated files permanently
const (
	Nanosecond time.Duration = 1
	Second = 1000000000*Nanosecond
	Hour = 3600*1000*1000*1000*Nanosecond
	OneDay	= 24*Hour
	OneWeek = 7*OneDay
)

//function to find time (in hours) since some recorded time.
func timeSinceSec(t time.Time)int{
   	timeSince := (time.Now().Sub(t)) //e.g. (time.Now().Sub(rfq.RFQDate))
    return int(timeSince/Second) //Second defined in const above
}


func timeSinceHours(t time.Time)int{
   	timeSince := (time.Now().Sub(t)) //e.g. (time.Now().Sub(rfq.RFQDate))
    return int(timeSince/Hour) //hour defined in const above
}

func timeSinceDays(t time.Time)int{
   	timeSince := (time.Now().Sub(t)) //e.g. (time.Now().Sub(rfq.RFQDate))
    return int(timeSince/OneDay) //One day is 24 hours defined in const above
}


const alphnum = "01234567890abcdefghijklmnopqrtsuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//Generating pseudo random strings for login tokens.
//See http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
//for a better, faster method. 

func RandStrings(n int) string{
	b:=make([]byte, n)
	for i:= range b{
		b[i] = alphnum[rand.Intn(len(alphnum))]
	}
return string(b)
}
 
//function to append a string to the front of an existing []string. 
func strArrAppZero(strArrIn []string, str string)([]string){
	temp := make([]string, len(strArrIn)+1)
	temp[0] = str
	if len(strArrIn) == 0{
		return temp
	}else{
		//using the built in append function should be better.
		for i := 0; i<len(strArrIn);i++{
			temp[i+1] = strArrIn[i]
		}
	}
return temp
}

//function to append a datastore key to a slice
func keyArrAppZero(keyArrIn []*datastore.Key, key *datastore.Key)([]*datastore.Key){
	temp := make([]*datastore.Key, len(keyArrIn)+1)
	temp[0] = key
	if len(keyArrIn) == 0{
		return temp
	}else{
		for i := 0; i<len(keyArrIn);i++{
			temp[i+1] = keyArrIn[i]
		}
	}
return temp
}

type RinsRespListSort []RInsRFQList

//Sorting function for Rins Premiums.
func (this RinsRespListSort) Len() int{
	return len(this)
}

func (this RinsRespListSort) Less(i,j int) bool{
	iPrem,erri := strconv.ParseFloat(this[i].Premium[0],64)
	if erri != nil{
		log.Printf("Problem converting string Premium to float in method Less")
	}
	jPrem,errj := strconv.ParseFloat(this[j].Premium[0],64)
	if errj != nil{
		log.Printf("Problem converting string Premium to float")
	} 	  
	return iPrem < jPrem  
}

func (this RinsRespListSort) Swap(i,j int) {
	this[i], this[j] = this[j], this[i]
}

type FltPremiums []float64

//Sorting function for Float Premiums to calculate median.
func (this FltPremiums) Len() int{
	return len(this)
}

func (this FltPremiums) Less(i,j int) bool{	
	return this[i] < this[j]  
}

func (this FltPremiums) Swap(i,j int) {
	this[i], this[j] = this[j], this[i]
}



//Mean
func mean(x []float64) float64{
	var sum float64
	for _, v := range x{
		sum += v 		
	}
	return sum/float64(len(x))
}

//Median using (n-1)
func median(x FltPremiums) (med float64, err error){
	num := len(x)
	if num < 3{
		med = 0
		err = medianerror
		return 
	}
	sort.Sort(x)
	if num % 2 != 0 { //odd
		med = x[(num-1)/2]
		return 
	}else{	//even number but more than 2 points
		m1 := x[(num/2)-1]
		m2 := x[num/2]
		med = mean([]float64{m1, m2})
		return 
	}
}
