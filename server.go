package main

import (
  "fmt"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "time"
  "github.com/julienschmidt/httprouter"
  "net/http"
  "encoding/json"
  "strings"
)



type LocReq struct {

    Name    string `json:"name"`
    Address string `json:"address"`
    City    string `json:"city"`
    State   string `json:"state"`
    Zip     string `json:"zip"`

}



type LocRes struct {
    ID    bson.ObjectId `json:"id" bson:"_id,omitempty"`
    Name  string `json:"name"`
    Address    string `json:"address"`
    City       string `json:"city"`
    State string `json:"state"`
    Zip   string `json:"zip"`
    Coordinate struct {
        Lat float64 `json:"lat"`
        Lng float64 `json:"lng"`
    } `json:"coordinate"`
}



type GeoAddress struct {
    Results []struct {
        AddressComponents []struct {
            LongName string `json:"long_name"`
            ShortName string `json:"short_name"`
            Types []string `json:"Types"`
        } `json:"address_components"`

        FormattedAddress string `json:"formatted_address"`

        Geometry struct {
            Location struct {
                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`

            } `json:"location"`
            LocationType string `json:"location_type"`

            Viewport struct {

                Norteast struct {

                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`

                } `json:"northeast"`

                Southwest struct {

                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`

                } `json:"southwest"`

            } `json:"viewport"`

        } `json:"geometry"`

        PlaceId string `json:"place_id"`
        Types string `json:"types"`

    } `json:"results"`

    Status string `json:"status"`

}

type AddressResponesGoogle struct {
    Results []struct {
        AddressComponents []struct {
            LongName string `json:"long_name"`
            ShortName string `json:"short_name"`
            Types []string `json:"Types"`
        } `json:"address_components"`

        FormattedAddress string `json:"formatted_address"`

        Geometry struct {
            Location struct {

                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`

            } `json:"location"`
            LocationType string `json:"location_type"`

            Viewport struct {

                Norteast struct {

                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`

                } `json:"northeast"`

                Southwest struct {

                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`

                } `json:"southwest"`

            } `json:"viewport"`

        } `json:"geometry"`

        PlaceId string `json:"place_id"`
        Types string `json:"types"`

    } `json:"results"`

    Status string `json:"status"`

}

var col *mgo.Collection
var locRes LocRes

const(
    timeout = time.Duration(time.Second*100)
)

func dbConnection() {

    db_uri := "mongodb://admin:admin@ds045064.mongolab.com:45064/tripadvisor"
    sess1,err1 := mgo.Dial(db_uri)

    if(err1 !=nil) {

        fmt.Printf("DB Connection error %v\n",err1)

    } else {

        sess1.SetSafe(&mgo.Safe{})
        col = sess1.DB("tripadvisor").C("location")

    }

}

func getGooglelocDetails( addr string) (geoAddress GeoAddress) {

    client := http.Client{Timeout: timeout}
    googleUrl := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s",addr)

    res1, err1 := client.Get(googleUrl)

    if(err1 !=nil) {

        fmt.Printf("Cannot fetch address details from google api %v\n",err1)

    }

    defer res1.Body.Close()

    dec := json.NewDecoder(res1.Body)

    err1 = dec.Decode(&geoAddress)

    if(err1 != nil)    {

        fmt.Errorf("Unable to decode json from Google: %v", err1)

    }

    return geoAddress


}

func getLoc(reswriter http.ResponseWriter, req *http.Request, par httprouter.Params) {

    

    id := bson.ObjectIdHex(par.ByName("locationID"))

    err := col.FindId(id).One(&locRes)

    if err != nil {

        fmt.Printf("got an error finding a doc %v\n")

    }   

    reswriter.Header().Set("Content-Type", "application/json; charset=UTF-8")

    reswriter.WriteHeader(200)

    json.NewEncoder(reswriter).Encode(locRes)

}

func addLoc(reswriter http.ResponseWriter, req *http.Request, par httprouter.Params) {



    var tempRequest LocReq

    dec := json.NewDecoder(req.Body)

    err := dec.Decode(&tempRequest)

    if(err!=nil)    {

        fmt.Errorf("Unable to decode json input %v", err)

    }

    addr := tempRequest.Address+" "+tempRequest.City+" "+tempRequest.State+" "+tempRequest.Zip

    addr = strings.Replace(addr," ","%20",-1)



    locDetails := getGooglelocDetails(addr)



    locRes.ID = bson.NewObjectId()

    locRes.Address= tempRequest.Address

    locRes.City=tempRequest.City

    locRes.Name=tempRequest.Name

    locRes.State=tempRequest.State

    locRes.Zip=tempRequest.Zip

    locRes.Coordinate.Lat=locDetails.Results[0].Geometry.Location.Lat

    locRes.Coordinate.Lng=locDetails.Results[0].Geometry.Location.Lng


    err = col.Insert(locRes)

    if err != nil {

        fmt.Printf("Unable to insert entry: %v\n", err)

    }



    err = col.FindId(locRes.ID).One(&locRes)

    if err != nil {

        fmt.Printf("unable to find entry %v\n")

    }   

    reswriter.Header().Set("Content-Type", "application/json; charset=UTF-8")

    reswriter.WriteHeader(201)

    json.NewEncoder(reswriter).Encode(locRes)

}

func updateLoc(reswriter http.ResponseWriter, req *http.Request, par httprouter.Params) {



    var tempRequest LocRes

    var locRes LocRes

    id := bson.ObjectIdHex(par.ByName("locationID"))

    err := col.FindId(id).One(&locRes)

    if err != nil {

        fmt.Printf("Unable to find entry %v\n")

    } 

    tempRequest.Name = locRes.Name

    tempRequest.Address = locRes.Address

    tempRequest.City = locRes.City

    tempRequest.State = locRes.State

    tempRequest.Zip = locRes.Zip

    decoder := json.NewDecoder(req.Body)

    err = decoder.Decode(&tempRequest)

    

    if(err!=nil)    {

        fmt.Errorf("Unable to decode json input: %v", err)

    }



    addr := tempRequest.Address+" "+tempRequest.City+" "+tempRequest.State+" "+tempRequest.Zip

    addr = strings.Replace(addr," ","%20",-1)

    locDetails := getGooglelocDetails(addr)

    tempRequest.Coordinate.Lat=locDetails.Results[0].Geometry.Location.Lat

    tempRequest.Coordinate.Lng=locDetails.Results[0].Geometry.Location.Lng

    err = col.UpdateId(id,tempRequest)

    if err != nil {

        fmt.Printf("Error in updating the entry %v\n")

    } 


    err = col.FindId(id).One(&locRes)

    if err != nil {

        fmt.Printf("Unable to find entry %v\n")

    }

    reswriter.Header().Set("Content-Type", "application/json; charset=UTF-8")

    reswriter.WriteHeader(201)

    json.NewEncoder(reswriter).Encode(locRes)

}

func deleteLoc(reswriter http.ResponseWriter, req *http.Request, par httprouter.Params) {



    id := bson.ObjectIdHex(par.ByName("locationID"))

    err := col.RemoveId(id)

    if err != nil {

        fmt.Printf("Unable to delete entry %v\n")

    }

    reswriter.WriteHeader(200)

}

func main() {

    mux := httprouter.New()

    mux.GET("/locations/:locationID", getLoc)

    mux.POST("/locations", addLoc)

    mux.PUT("/locations/:locationID", updateLoc)

    mux.DELETE("/locations/:locationID", deleteLoc)

    server := http.Server{

            Addr:        "0.0.0.0:8080",

            Handler: mux,

    }

    dbConnection()

    server.ListenAndServe()

}

