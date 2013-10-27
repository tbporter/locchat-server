package main

import (
	"container/list"
	"encoding/json"
    "fmt"
    "net/http"
    "sync"
    "math"
)

//Set up them http handlers
func main(){
	messages.L = list.New()
	http.HandleFunc("/chat", chatHandler)
	http.HandleFunc("/poll", pollHandler)    
	http.ListenAndServe(":8080", nil)
}


//I implemented a list with a lock instead of database! time constraints!
type message_list struct {
	Mu sync.RWMutex
	L *list.List
}

var messages message_list

func addMessage(msg chat_msg_struct){
	messages.Mu.Lock()
	defer messages.Mu.Unlock()
	messages.L.PushFront(msg)
	if messages.L.Len()>100{
		messages.L.Remove(messages.L.Back())
	}
}

//This goes through the list until it finds n messages in range...
//but will go through the whole list if it can't find n, I'll implement the database later
func getMessage(n int, req chat_msg_req_struct) []chat_msg_struct{
	var ms []chat_msg_struct

	messages.Mu.RLock()
	defer messages.Mu.RUnlock()
	var curMsg chat_msg_struct;
	for i :=  messages.L.Front(); i != nil; i = i.Next(){
		if n<=0 {
			break
		}
		curMsg = i.Value.(chat_msg_struct)
		if distInRange(curMsg.Lat,curMsg.Lon,req.Lat,req.Lon,req.Dist){
			ms = append(ms, curMsg)
			n--;
		}
	}

	return ms
}

//Recieves chat_msg_struct from html and stores it
func chatHandler(w http.ResponseWriter, r *http.Request){
    fmt.Println("handling chat")
    decoder := json.NewDecoder(r.Body)
    var m chat_msg_struct
    err := decoder.Decode(&m)
    if err != nil {
    	fmt.Println(err)
 		return;
    }
    addMessage(m)
    fmt.Printf("[%f,%f] %s: %s\n",m.Lat,m.Lon,m.Usr,m.Msg)
}

//recieves a chat_msg_req_struct from html, responds with a chat_msgs_struct
func pollHandler(w http.ResponseWriter, r *http.Request){
	fmt.Println("handling poll")

	decoder := json.NewDecoder(r.Body)
    var m chat_msg_req_struct
    err := decoder.Decode(&m)
    if err != nil {
    	fmt.Println(err)
 		return;
    }

	enc := json.NewEncoder(w)
	ms := chat_msgs_struct{getMessage(25,m)}
	enc.Encode(&ms)
}


//This is a really rough estimate of latlon distances, we don't care too much about specifics so lets not waste time calculating
func distInRange(lat1 float64,lon1 float64,lat2 float64,lon2 float64, dist int) bool {
	const EST_DIST = 60 //estimated dist in miles between degrees, biased against people on the poles, those guys are jerks anway
	//converting dist which is from 0 to 100, to a better scale
	if dist<1{
		dist = 1
	}
	maxdist := math.Pow(float64(dist)/13,3)
	if math.Abs(lat1-lat2) <= maxdist && math.Abs(lon1-lon2) <= maxdist{
		return true
	}
	fmt.Printf("[%f,%f] not in range of [%f,%f] with dist of %f\n",lat1,lon1,lat2,lon2,maxdist)
	return false
}

//JSON structs
type chat_msg_struct struct {
	Msg string
	Usr string
	Time int64
	Lat float64
	Lon float64
}

type chat_msg_req_struct struct {
	Usr string
	Lat float64
	Lon float64
	Dist int
}

type chat_msgs_struct struct {
	Msgs []chat_msg_struct
}