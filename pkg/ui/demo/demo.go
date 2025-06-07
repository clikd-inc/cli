package demo

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"clikd/pkg/ui"
	"clikd/pkg/ui/components"
)

// RunDemo führt die Demo der UI-Komponenten aus
func RunDemo() {
	fmt.Println(ui.RenderLogo())
	fmt.Println(ui.H1.Render("UI-Komponenten Demo"))
	fmt.Println()

	demoMenu()
}

// demoMenu zeigt ein Auswahlmenü für die verschiedenen Demos
func demoMenu() {
	for {
		items := []components.SelectItem{
			{Title: "Select-Komponente", Description: "Ein einfaches Auswahlmenü", Value: "select"},
			{Title: "Multiselect-Komponente", Description: "Mehrfachauswahl mit Checkboxen", Value: "multiselect"},
			{Title: "Confirm-Komponente", Description: "Ja/Nein-Entscheidungen", Value: "confirm"},
			{Title: "Input-Komponente", Description: "Texteingabe", Value: "input"},
			{Title: "Progress-Komponente", Description: "Fortschrittsbalken", Value: "progress"},
			{Title: "List-Komponente", Description: "Darstellung von Listeneinträgen", Value: "list"},
			{Title: "Styling-Beispiele", Description: "Verfügbare Stile und Farben", Value: "styles"},
			{Title: "Beenden", Description: "Demo beenden", Value: "exit"},
		}

		selected := components.RunSelect("Wähle eine Demo-Komponente", items)
		if selected == nil {
			break
		}

		switch selected.Value {
		case "select":
			demoSelect()
		case "multiselect":
			demoMultiselect()
		case "confirm":
			demoConfirm()
		case "input":
			demoInput()
		case "progress":
			demoProgress()
		case "list":
			demoList()
		case "styles":
			demoStyles()
		case "exit":
			fmt.Println(ui.SuccessText("Demo beendet."))
			return
		}

		// Pause zwischen den Demos
		fmt.Println(ui.SubtleText.Render("Drücke Enter, um fortzufahren..."))
		fmt.Scanln()
	}
}

// demoSelect demonstriert die Select-Komponente
func demoSelect() {
	items := []components.SelectItem{
		{Title: "Option 1", Description: "Die erste Option", Value: 1},
		{Title: "Option 2", Description: "Die zweite Option", Value: 2},
		{Title: "Option 3", Description: "Die dritte Option", Value: 3},
	}

	fmt.Println(ui.H2.Render("Select-Komponente Demo"))
	fmt.Println(ui.NormalText.Render("Diese Komponente bietet ein interaktives Auswahlmenü."))
	fmt.Println()

	selected := components.RunSelect("Wähle eine Option", items)
	if selected == nil {
		fmt.Println(ui.WarningText("Keine Auswahl getroffen (abgebrochen)."))
	} else {
		fmt.Printf(ui.SuccessText("Ausgewählt: %s (Wert: %v)\n"), selected.Title, selected.Value)
	}
}

// demoMultiselect demonstriert die Multiselect-Komponente
func demoMultiselect() {
	items := []components.MultiselectItem{
		{Title: "Option 1", Description: "Die erste Option", Value: 1},
		{Title: "Option 2", Description: "Die zweite Option", Value: 2, Selected: true},
		{Title: "Option 3", Description: "Die dritte Option", Value: 3},
		{Title: "Option 4", Description: "Die vierte Option", Value: 4},
		{Title: "Option 5", Description: "Die fünfte Option", Value: 5},
	}

	fmt.Println(ui.H2.Render("Multiselect-Komponente Demo"))
	fmt.Println(ui.NormalText.Render("Diese Komponente ermöglicht die Auswahl mehrerer Optionen."))
	fmt.Println(ui.SubtleText.Render("Tipp: Verwende Leertaste zum Auswählen/Abwählen und 'a' um alle auszuwählen/abzuwählen."))
	fmt.Println()

	selected := components.RunMultiselect("Wähle mehrere Optionen", "Du kannst beliebig viele Optionen auswählen.", items)
	if selected == nil {
		fmt.Println(ui.WarningText("Keine Auswahl getroffen (abgebrochen)."))
	} else {
		fmt.Println(ui.SuccessText(fmt.Sprintf("Ausgewählte Optionen: %d", len(selected))))
		for i, item := range selected {
			fmt.Printf("  %d. %s (Wert: %v)\n", i+1, item.Title, item.Value)
		}
	}

	// Demo mit maximaler Auswahl
	fmt.Println()
	fmt.Println(ui.H2.Render("Multiselect mit maximaler Auswahl"))
	maxItems := components.RunMultiselectWithMaxSelected("Wähle bis zu 2 Optionen", "Du kannst maximal 2 Optionen auswählen.", items, 2)
	if maxItems == nil {
		fmt.Println(ui.WarningText("Keine Auswahl getroffen (abgebrochen)."))
	} else {
		fmt.Println(ui.SuccessText(fmt.Sprintf("Ausgewählte Optionen: %d", len(maxItems))))
		for i, item := range maxItems {
			fmt.Printf("  %d. %s (Wert: %v)\n", i+1, item.Title, item.Value)
		}
	}
}

// demoConfirm demonstriert die Confirm-Komponente
func demoConfirm() {
	fmt.Println(ui.H2.Render("Confirm-Komponente Demo"))
	fmt.Println(ui.NormalText.Render("Diese Komponente ermöglicht Ja/Nein-Entscheidungen."))
	fmt.Println()

	// Einfache Bestätigung
	result := components.Confirm(
		"Bestätigung",
		"Möchtest du wirklich fortfahren? Diese Aktion kann nicht rückgängig gemacht werden.",
	)

	if result {
		fmt.Println(ui.SuccessText("Du hast mit 'Ja' geantwortet."))
	} else {
		fmt.Println(ui.WarningText("Du hast mit 'Nein' geantwortet oder abgebrochen."))
	}

	// Bestätigung mit Standardwert
	fmt.Println()
	fmt.Println(ui.H2.Render("Confirm mit Standardwert"))
	resultWithDefault := components.ConfirmWithDefault(
		"Einstellungen speichern",
		"Möchtest du deine Einstellungen speichern? Drücke ESC für den Standardwert.",
		true, // Standardwert: Ja
	)

	if resultWithDefault {
		fmt.Println(ui.SuccessText("Einstellungen wurden gespeichert (oder Standardwert 'Ja' verwendet)."))
	} else {
		fmt.Println(ui.WarningText("Einstellungen wurden nicht gespeichert."))
	}
}

// demoInput demonstriert die Input-Komponente
func demoInput() {
	fmt.Println(ui.H2.Render("Input-Komponente Demo"))
	fmt.Println(ui.NormalText.Render("Diese Komponente ermöglicht Texteingaben."))
	fmt.Println()

	// Einfache Eingabe
	name := components.RunInput(
		"Wie heißt du?",
		"Bitte gib deinen Namen ein:",
		"Max Mustermann",
	)

	if name == "" {
		fmt.Println(ui.WarningText("Keine Eingabe (abgebrochen)."))
	} else {
		fmt.Println(ui.SuccessText(fmt.Sprintf("Hallo, %s!", name)))
	}

	// Eingabe mit Standardwert
	fmt.Println()
	fmt.Println(ui.H2.Render("Input mit Standardwert"))
	email := components.RunInputWithDefault(
		"E-Mail-Adresse",
		"Bitte gib deine E-Mail-Adresse ein (oder ESC für den Standardwert):",
		"email@example.com",
		"default@example.com",
	)

	fmt.Println(ui.SuccessText(fmt.Sprintf("E-Mail-Adresse: %s", email)))
}

// demoProgress demonstriert die Progress-Komponente
func demoProgress() {
	fmt.Println(ui.H2.Render("Progress-Komponente Demo"))
	fmt.Println(ui.NormalText.Render("Diese Komponente zeigt den Fortschritt einer Operation an."))
	fmt.Println()

	// Prozentbasierter Fortschritt
	fmt.Println(ui.SubtleText.Render("Einfacher Fortschrittsbalken:"))
	components.RunProgress(
		"Daten werden verarbeitet",
		"Bitte warten, während die Daten verarbeitet werden...",
		func(setPercent func(float64), setDone func()) {
			// Simuliere eine Operation mit zufälligen Fortschritten
			for i := 0.0; i <= 1.0; i += 0.1 {
				time.Sleep(time.Millisecond * time.Duration(300+rand.Intn(300)))
				setPercent(i)
			}
			setDone()
		},
	)

	// Wertbasierter Fortschritt
	fmt.Println()
	fmt.Println(ui.SubtleText.Render("Fortschrittsbalken mit Werten:"))
	components.RunProgressWithValues(
		"Dateien werden hochgeladen",
		"Bitte warten, während die Dateien hochgeladen werden...",
		10, // Maximaler Wert: 10 Dateien
		"Dateien",
		func(setValue func(int), setDone func()) {
			// Simuliere einen Datei-Upload
			for i := 1; i <= 10; i++ {
				time.Sleep(time.Millisecond * time.Duration(300+rand.Intn(300)))
				setValue(i)
			}
			setDone()
		},
	)
}

// demoList demonstriert die List-Komponente
func demoList() {
	fmt.Println(ui.H2.Render("List-Komponente Demo"))
	fmt.Println(ui.NormalText.Render("Diese Komponente zeigt eine formatierte Liste von Elementen an."))
	fmt.Println()

	// Erstelle eine Liste von Elementen
	items := []components.ListItem{
		components.CreateListItemWithStatus("Task 1", "Dies ist die Beschreibung für Task 1", "done"),
		components.CreateListItemWithStatus("Task 2", "Dies ist die Beschreibung für Task 2", "in progress"),
		components.CreateListItemWithStatus("Task 3", "Dies ist die Beschreibung für Task 3", "pending"),
		components.CreateListItemWithStatus("Task 4", "Dies ist die Beschreibung für Task 4", "error"),
	}

	// Füge Tags und Metadaten hinzu
	items[0].Tags = []string{"wichtig", "abgeschlossen"}
	items[1].Tags = []string{"in Bearbeitung", "dringend"}
	items[2].Metadata = map[string]string{
		"Fällig bis": "Morgen",
		"Priorität":  "Hoch",
	}
	items[3].Metadata = map[string]string{
		"Fehlermeldung": "Konnte nicht verbinden",
		"Versuch":       "3 von 5",
	}

	// Füge mehr Elemente hinzu, um Paginierung zu demonstrieren
	for i := 5; i <= 15; i++ {
		status := "pending"
		if i%3 == 0 {
			status = "done"
		} else if i%3 == 1 {
			status = "in progress"
		}
		items = append(items, components.ListItem{
			Title:       fmt.Sprintf("Task %d", i),
			Description: fmt.Sprintf("Dies ist die Beschreibung für Task %d", i),
			Status:      status,
			Tags:        []string{fmt.Sprintf("tag-%d", i)},
		})
	}

	// Zeige die Liste an
	components.ShowFormattedList(
		"Aufgabenliste",
		"Dies ist eine Liste aller Aufgaben. Verwende die Pfeiltasten zum Navigieren.",
		items,
		true, // Tags anzeigen
		true, // Status anzeigen
		5,    // 5 Einträge pro Seite
	)
}

// demoStyles demonstriert die verfügbaren Stile und Farben
func demoStyles() {
	fmt.Println(ui.H2.Render("Styles-Demo"))
	fmt.Println(ui.NormalText.Render("Diese Demo zeigt die verfügbaren Stile und Farben."))
	fmt.Println()

	// Text-Stile
	fmt.Println(ui.H2.Render("Text-Stile"))
	fmt.Println(ui.NormalText.Render("Normaler Text"))
	fmt.Println(ui.BoldText.Render("Fettgedruckter Text"))
	fmt.Println(ui.SubtleText.Render("Dezenter Text"))
	fmt.Println(ui.Selected.Render("Ausgewählter Text"))
	fmt.Println(ui.Highlight.Render("Hervorgehobener Text"))
	fmt.Println()

	// Überschriften
	fmt.Println(ui.H1.Render("Überschrift 1"))
	fmt.Println(ui.H2.Render("Überschrift 2"))
	fmt.Println()

	// Status-Stile
	fmt.Println(ui.H2.Render("Status-Stile"))
	fmt.Println(ui.SuccessText("Erfolg"))
	fmt.Println(ui.ErrorText("Fehler"))
	fmt.Println(ui.WarningText("Warnung"))
	fmt.Println(ui.InfoText("Information"))
	fmt.Println()

	// Box-Stil
	fmt.Println(ui.H2.Render("Box-Stil"))
	fmt.Println(ui.Box("Dies ist ein Text in einer Box.\nDie Box kann mehrere Zeilen enthalten und passt sich an den Inhalt an."))
	fmt.Println()

	// Sektionen
	fmt.Println(ui.SectionTitle("Abschnittstitel"))
	fmt.Println(ui.NormalText.Render("Ein Abschnitt mit einem Titel."))
	fmt.Println()

	// Zentrierter Text
	fmt.Println(ui.H2.Render("Zentrierter Text"))
	width, _, _ := term()
	fmt.Println(ui.CenterText("Dieser Text ist zentriert", width))
	fmt.Println()
}

// term gibt die Terminalabmessungen zurück
func term() (width, height int, err error) {
	defer func() {
		if r := recover(); r != nil {
			width = 80
			height = 24
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// Fallback-Werte
	width = 80
	height = 24

	// Versuche, die tatsächliche Terminalgröße zu erhalten
	if w, h, err := terminalSize(); err == nil {
		width = w
		height = h
	}

	return width, height, nil
}

// terminalSize gibt die Terminalgröße zurück (falls verfügbar)
func terminalSize() (width, height int, err error) {
	// os.Stdout ist bereits ein *os.File, keine Type-Assertion nötig
	if width, height, err = getTerminalSize(os.Stdout.Fd()); err != nil {
		return 80, 24, err
	}
	return width, height, nil
}

// getTerminalSize wird in plattformspezifischen Dateien implementiert:
// - term_unix.go für Linux/macOS
// - term_windows.go für Windows
