package main

import (
	"fmt"
	"log"
	"github.com/gocolly/colly"
	"net/http"
	"net/url"
	"html/template"
	"os"
	"encoding/csv"
	"path/filepath"
)
type Link struct {
	Tittle string 
	Href string
}
type SearchData struct {
	Text string
	Query string
	Link_list []Link
}
//Při načtení stránky
func index(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("index.html")  
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
	}
	err = temp.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
	}
}
//Při zadaném vstupu
func search(w http.ResponseWriter, r *http.Request)  {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	search_text := r.FormValue("search")//načte vstup od uživatele
	search_text_encoded := url.QueryEscape(search_text)//převede text do URL formátu
	links, err := scrape(search_text_encoded)//provede request na googlu
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
	}
	temp, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
	}
	data := SearchData{search_text, search_text_encoded, links}
	err = temp.Execute(w, data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
	}
	err = writeData(search_text_encoded, &links)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
	}
}
func download(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename") + ".csv"
  directory := filepath.Join("download", filename)
	// open file (check if exists)
	_, err := os.Open(directory)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(directory))
  w.Header().Set("Content-Type", "application/octet-stream") // Set binary stream type
  // Serve the file
  http.ServeFile(w, r, directory)
}

func delete_(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename") + ".csv"
	fmt.Println("deleting: "+filename)
  directory := filepath.Join("download", filename)
	// open file (check if exists)
	_, err := os.Open(directory)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "<h1>Internal server error</h1>", http.StatusInternalServerError)
		return
	}
	os.Remove(directory)
}

//zapiš data do tabulky
func writeData(filename string, links *[]Link) error{
	file, err := os.Create("download/" + filename + ".csv")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	err = writer.Write([]string{"Tittle", "Link"})
	for _, link := range *links {
		err = writer.Write([]string{link.Tittle, link.Href})
	}
	return err
}

//pošle request na google
func scrape(input string) ([]Link, error){
	query := "https://www.google.com/search?q=" + input
	var links []Link //array se všemi linky
	var error_ error 
	c := colly.NewCollector(
		colly.AllowedDomains("www.google.com"),
	)
	//abychom se tvářili jako člověk a nebyli blokováni
	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0"
	c.OnError(func(_ *colly.Response, err error) { //v případě problému s requestem
        fmt.Printf("Něco se pokazilo: %v", err)
				error_ = err
    })
	c.OnHTML("#search", func(e *colly.HTMLElement) {
		e.ForEach(".yuRUbf", func(i int, h *colly.HTMLElement) { //najde všechny odkazy
			h3 := h.ChildText("h3") //tittle stránky
			a := h.ChildAttr("a", "href") //odkaz na stránku
			l := Link{h3, a}
			links = append(links, l)
		})
	})
	c.Visit(query)
	return links, error_ 
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/search", search)
	http.HandleFunc("/download/{filename}", download)
	http.HandleFunc("/delete/{filename}", delete_)
	fmt.Println("Ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
