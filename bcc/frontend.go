package bcc

import (
	//"encoding/json"
	"fmt"
	"net/http"
	//"os"
	"html/template"
)

// Data structure to hold information from the JSON fil
type WalletDataParsable struct {
	WalletID string `json:"WalletID"`
	Data WalletData
}
func Server(db *DBmap) {
	// Read JSON data from a file
	var data []WalletDataParsable
	for key := range *db {
		data = append(data, WalletDataParsable{
			WalletID: key,
			Data: (*db)[key],
		})
	}

	// Handle requests to the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create a simple HTML template
		tmpl := `<html>
					<head>
						<title>Simple Webpage</title>
					</head>
					<body>
						<h1>Table from JSON</h1>
						<table border="1">
							<tr>
								<th>WalletID</th>
								<th>Balance</th>
							</tr>
							{{range .}}
								<tr>
									<td>{{.WalletID}}</td>
									<td>{{.Data.Balance}}</td>
								</tr>
							{{end}}
						</table>
					</body>
				</html>`

		// Parse the HTML template
		t, err := template.New("webpage").Parse(tmpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute the template with the data from the JSON file
		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Start the HTTP server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
