// Plot Helper
// Dependency References:
// https://pkg.go.dev/gonum.org/v1/plot
// https://pkg.go.dev/gonum.org/v1/plot@v0.14.0/plotter
// https://www.w3schools.com/css/css_colors_rgb.asp

package helpers

import (
	"fmt"
	"image/color"
	"log"
	"ppolls2024/global"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func plotOneState(state string, endDateArray []string, demPctArray, gopPctArray, otherPctArray []float64) {
	glob := global.GetGlobalRef()
	RED := color.NRGBA{R: 255, A: 255}
	//GREEN := color.NRGBA{G: 255, A: 255}
	BLUE := color.NRGBA{B: 255, A: 255}
	GREY := color.RGBA{180, 180, 180, 255}
	BLACK := color.Black

	plt := plot.New()
	plt.Title.Text = fmt.Sprintf("%s Polling", state)
	plt.X.Tick.Marker = plot.TimeTicks{Format: "Jan 02"}
	plt.Y.Tick.Length.Dots(5.0)
	plt.Y.Label.Text = "Voter %"
	plt.Add(plotter.NewGrid())

	countPoints := 0

	linePoints := func(dateArray []string, dependent []float64) plotter.XYs {
		arraySizes := len(dateArray)
		pts := make(plotter.XYs, arraySizes)
		for ix := range pts {
			layout := string(time.RFC3339[:10])
			tm, err := time.Parse(layout, dateArray[ix])
			if err != nil {
				log.Fatalf("plotOneState::linePoints: time.Parse(%s, %s) failed, reason: %s\n", layout, dateArray[ix], err.Error())
			}
			if tm.Before(glob.DateThreshold) {
				continue
			}
			countPoints++
			timeInt64 := time.Date(tm.Year(), tm.Month(), tm.Day(), 12, 30, 30, 0, time.UTC).Unix()
			pts[ix].X = float64(timeInt64)
			pts[ix].Y = dependent[ix]
		}
		return pts
	}

	data := linePoints(endDateArray, demPctArray)
	if countPoints < 1 {
		return
	}

	log.Printf("State plot for %s .....\n", state)

	line, points, err := plotter.NewLinePoints(data)
	if err != nil {
		log.Fatalf("plotOneState: internal error diagnosed in plotter.NewLinePoints(dem), reason: %s\n" + err.Error())
	}
	line.Color = BLUE
	line.Width = 2
	points.Shape = draw.CircleGlyph{}
	points.Color = BLACK
	plt.Add(line, points)

	data = linePoints(endDateArray, gopPctArray)
	line, points, err = plotter.NewLinePoints(data)
	if err != nil {
		log.Fatalf("plotOneState: internal error diagnosed in plotter.NewLinePoints(gop), reason: %s\n" + err.Error())
	}
	line.Color = RED
	line.Width = 2
	points.Shape = draw.CircleGlyph{}
	points.Color = BLACK
	plt.Add(line, points)

	data = linePoints(endDateArray, otherPctArray)
	line, points, err = plotter.NewLinePoints(data)
	if err != nil {
		log.Fatalf("plotOneState: internal error diagnosed in plotter.NewLinePoints(other), reason: %s\n" + err.Error())
	}
	line.Color = GREY
	line.Width = 2
	points.Shape = draw.CircleGlyph{}
	points.Color = BLACK
	plt.Add(line, points)

	err = plt.Save(vg.Length(glob.PlotWidth)*vg.Centimeter,
		vg.Length(glob.PlotHeight)*vg.Centimeter,
		fmt.Sprintf("%s/%s.png", glob.DirPlots, state))
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func Plodder() {
	glob := global.GetGlobalRef()
	var stateECV ECVote
	for _, stateECV = range stateECVTable {
		// For the given state, query from the most recent to the least recent polling.
		sqlText := fmt.Sprintf("SELECT end_date, pct_dem, pct_gop FROM history WHERE state = '%s' ORDER BY end_date DESC",
			stateECV.state)
		rows := sqlQuery(sqlText)

		var query dbparams
		var endDateArray []string
		var demPctArray []float64
		var gopPctArray []float64
		var otherPctArray []float64
		counter := 0
		for rows.Next() {
			counter += 1
			err := rows.Scan(&query.endDate, &query.pctDem, &query.pctGop)
			if err != nil {
				log.Fatalf("Plodder: rows.Scan failed, row count: %d, reason: %s\n", counter, err.Error())
			}
			endDateArray = append(endDateArray, query.endDate)
			demPctArray = append(demPctArray, query.pctDem)
			gopPctArray = append(gopPctArray, query.pctGop)
			curOtherPct := CalcOther(query.pctDem, query.pctGop)
			otherPctArray = append(otherPctArray, curOtherPct)
			if counter >= glob.PollHistoryLimit {
				break
			}
		}
		plotOneState(stateECV.state, endDateArray, demPctArray, gopPctArray, otherPctArray)
	}
	log.Println("State plots completed")
}
