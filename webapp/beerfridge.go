package beerfridge

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
)

func init() {
	http.HandleFunc("/", status)
	http.HandleFunc("/store", store)
	http.HandleFunc("/temp", temp)
}

type SensorData struct {
	Date     time.Time
	Temp_f   float64
	Keg1_ml  float64
	Keg2_ml  float64
	FridgeOn bool
}

func sensorKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "SensorData", "default_series", 0, nil)
}

const statusHTML = `
<html>
  <head>
    <meta name=viewport content="width=device-width, initial-scale=1">
    <meta http-equiv="refresh" content="5">
    <script type="text/javascript" src="https://www.google.com/jsapi"></script>
    <script type="text/javascript">
      google.load("visualization", "1", {packages:["gauge"]});
      google.setOnLoadCallback(drawChart);

      function drawChart() {
        drawKeg1();
        drawKeg2();
        drawTemp();
      }

      function drawKeg1() {
        var data = google.visualization.arrayToDataTable([
          ['Label', 'Value'],
          ['Keg 1 (ml)', {{.Keg1_ml}}],
        ]);

        var options = {
          width: 360, height: 180, max: 19000,
          redFrom: 0, redTo: 2580,
          yellowFrom: 2580, yellowTo: 5700,
          minorTicks: 5
        };

        var c = new google.visualization.Gauge(document.getElementById('keg1'));
        c.draw(data, options);
      }

      function drawKeg2() {
        var data = google.visualization.arrayToDataTable([
          ['Label', 'Value'],
          ['Keg 2 (ml)', {{.Keg2_ml}}],
        ]);

        var options = {
          width: 360, height: 180, max: 19000,
          redFrom: 0, redTo: 2580,
          yellowFrom: 2580, yellowTo: 5700,
          minorTicks: 5
        };

        var c = new google.visualization.Gauge(document.getElementById('keg2'));
        c.draw(data, options);
      }

      function drawTemp() {
        var data = google.visualization.arrayToDataTable([
          ['Label', 'Value'],
          ['Temp F', {{.Temp_f}}],
        ]);

        var options = {
          width: 180, height: 180, max: 75,
          redFrom: 55, redTo: 75,
          yellowFrom: 45, yellowTo: 55,
          greenFrom: 34, greenTo: 45,
          minorTicks: 5
        };

        var c = new google.visualization.Gauge(document.getElementById('temp'));
        c.draw(data, options);
      }
    </script>
  </head>
  <body>
    <div id="keg1" style="width: 180px; height: 180px; float: left"></div>
    <div id="keg2" style="width: 180px; height: 180px; float: left"></div>
		<a href="/temp" style="color: inherit">
      <div id="temp" style="width: 180px; height: 180px; float: left"></div>
      {{if .FridgeOn}}
      <div id="fridge" style="width: 180px; height: 180px; background-color: #d9ead3; text-align: center; vertical-align: middle; float: left">
        Fridge On
      </div>
      {{else}}
      <div id="fridge" style="width: 180px; height: 180px; background-color: #f4cccc; text-align: center; vertical-align: middle; float: left">
        Fridge Off
      </div>
      {{end}}
    </a>
		<div style="clear: both; float: left"><br>Last Update: {{.Date}}<div>
  </body>
</html>
`

var statusTemplate = template.Must(template.New("status").Parse(statusHTML))

func status(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("SensorData").Ancestor(sensorKey(c)).Order("-Date").Limit(1)
	data := make([]SensorData, 0, 10)
	if _, err := q.GetAll(c, &data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := statusTemplate.Execute(w, data[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func store(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	data := SensorData{Date: time.Now()}

	if r.FormValue("secret") != "beerisgood" {
		http.Error(w, "sorry", http.StatusInternalServerError)
		return
	}

	if v, err := strconv.ParseFloat(r.FormValue("temp_f"), 64); err == nil {
		data.Temp_f = v
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if v, err := strconv.ParseFloat(r.FormValue("keg1"), 64); err == nil {
		data.Keg1_ml = v
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if v, err := strconv.ParseFloat(r.FormValue("keg2"), 64); err == nil {
		data.Keg2_ml = v
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if v, err := strconv.ParseBool(r.FormValue("fridge_on")); err == nil {
		data.FridgeOn = v
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := datastore.NewIncompleteKey(c, "SensorData", sensorKey(c))
	_, err := datastore.Put(c, key, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, data)
}

const tempHTML = `
<html>
  <head>
    <meta name=viewport content="width=device-width, initial-scale=1">
    <script type="text/javascript" src="https://www.google.com/jsapi"></script>
    <script type="text/javascript">
      google.load("visualization", "1.1", {packages:["line"]});
      google.setOnLoadCallback(drawChart);

      function drawChart() {
        var data = google.visualization.arrayToDataTable([
          ['Minutes Ago', 'Temp'],
          {{range $index, $element := .}}
          [{{formatX $index}}, {{$element.Temp_f}}],
          {{end}}
        ]);

        var options = {
          chart: { title: 'Fridge Temp (F)' },
          width: 700,
          height: 300,
        };

        var c = new google.charts.Line(document.getElementById('temp'));
        c.draw(data, options);
      }
    </script>
  </head>
  <body>
    <div id="temp" style="width: 700px; height: 300px; float: left"></div>
    {{$d := index . 0}}
    <div style="clear: both"><br>Last Update: {{$d.Date}}<div>
  </body>
</html>
`

var tempTemplate = template.Must(template.New("temp").Funcs(template.FuncMap{"formatX": formatX}).Parse(tempHTML))

func formatX(idx int) float64 {
	return (float64(idx) * -5) / 60
}

func temp(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("SensorData").Ancestor(sensorKey(c)).Order("-Date").Limit(1000)
	data := make([]SensorData, 0, 1000)
	if _, err := q.GetAll(c, &data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// reverse the list
	//for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
	//	data[i], data[j] = data[j], data[i]
	//}

	err := tempTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
