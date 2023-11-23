package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp string
	Status    string
	Action    string
	Package   string
}

type LogEntries []LogEntry

func (l LogEntries) Len() int      { return len(l) }
func (l LogEntries) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l LogEntries) Less(i, j int) bool {
	timeFormat := "2006-01-02 15:04:05"
	t1, _ := time.Parse(timeFormat, l[i].Timestamp)
	t2, _ := time.Parse(timeFormat, l[j].Timestamp)

	// Comparez d'abord par année
	if t1.Year() != t2.Year() {
		return t1.Year() > t2.Year()
	}

	// Si les années sont les mêmes, comparez par mois
	if t1.Month() != t2.Month() {
		return t1.Month() > t2.Month()
	}

	// Si les années et les mois sont les mêmes, comparez par jour
	return t1.Day() > t2.Day()
}

func main() {
	startTime := time.Now() // Enregistrez le temps de début

	file, err := os.Open("/var/log/dpkg.log")
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier:", err)
		return
	}
	defer file.Close()

	var logEntries LogEntries

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			entry := LogEntry{
				Timestamp: strings.Join(parts[:3], " "),
				Status:    parts[3],
				Action:    parts[4],
				Package:   strings.Join(parts[5:], " "),
			}
			logEntries = append(logEntries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Erreur lors de la lecture du fichier:", err)
		return
	}

	// Triez les entrées de journal
	sort.Sort(logEntries)

	// Créez une carte pour stocker les logs par jour
	logsByDay := make(map[string]LogEntries)

	// Remplissez la carte avec les logs triés par jour
	for _, entry := range logEntries {
		day := entry.Timestamp[:10]
		logsByDay[day] = append(logsByDay[day], entry)
	}

	// Limitez les jours aux 3 derniers jours
	var lastThreeDays []string
	for day := range logsByDay {
		lastThreeDays = append(lastThreeDays, day)
	}

	// Créer un gestionnaire (handler) pour afficher les logs par jour
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ajouter le style CSS
		fmt.Fprint(w, `
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 20px;
			}

			h1 {
				text-align: center;
			}

			table {
				width: 100%;
				border-collapse: collapse;
				margin-top: 20px;
			}

			th, td {
				border: 1px solid #ddd;
				padding: 8px;
				text-align: left;
			}

			th {
				background-color: #f2f2f2;
			}

			form {
				display: inline-block;
				padding: 10px;
				border: 1px solid #ccc;
				border-radius: 5px;
			}

			select {
				padding: 5px;
				font-size: 14px;
			}

			input {
				padding: 8px;
				font-size: 14px;
				background-color: #4CAF50;
				color: white;
				border: none;
				border-radius: 5px;
				cursor: pointer;
			}
		</style>
	`)

		// Ajouter le titre H1
		fmt.Fprint(w, "<h1>LOG DPKG error detector</h1>")

		// Affichez la liste déroulante
		fmt.Fprint(w, "<form action=\"/logs\" method=\"get\">")
		fmt.Fprint(w, "<label for=\"days\">Choisissez un jour:</label>")
		fmt.Fprint(w, "<select id=\"days\" name=\"day\">")
		for _, day := range lastThreeDays {
			fmt.Fprintf(w, "<option value=\"%s\">%s</option>", day, day)
		}
		fmt.Fprint(w, "</select>")
		fmt.Fprint(w, "<input type=\"submit\" value=\"Afficher\"></form>")

		// Affichez les logs du jour sélectionné
		selectedDay := r.FormValue("day")
		if selectedDay != "" {
			// Déclarer et calculer elapsedTime ici
			elapsedTime := time.Since(startTime)
			fmt.Fprintf(w, "Nombre total de lignes traitées : %d<br>", len(logEntries))
			fmt.Fprintf(w, "Temps total de traitement : %s<br><br>", elapsedTime)

			fmt.Fprintf(w, "<h2>Logs pour le %s :</h2>", selectedDay)
			fmt.Fprint(w, "<table>")
			fmt.Fprint(w, "<tr><th>Timestamp</th><th>Status</th><th>Action</th><th>Package</th></tr>")
			for _, entry := range logsByDay[selectedDay] {
				fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
					entry.Timestamp, entry.Status, entry.Action, entry.Package)
			}
			fmt.Fprint(w, "</table>")
		}
	})

	// Lancer le serveur sur le port 8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur lors du démarrage du serveur:", err)
	}
}
