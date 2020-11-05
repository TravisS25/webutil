package chart

/////////////////////////////////////////////////////////////////////////////
// This file contains type safe structs for various objects used in
// the chart.js library
//
// The structs are based on the type definitions from the
// npm package @types/chart.js (https://www.npmjs.com/package/@types/chart.js)
/////////////////////////////////////////////////////////////////////////////

// // ChartType enums
// const(
// 	LineChartType ChartType = "line"
// 	BarChartType = "bar"
// 	HorizontalBarChartType = "horizontalBar"
// 	RadarChartType = "radar"
// 	DoughnutChartType = "doughnut"
// 	PolarAreaChartType = "polarArea"
// 	BubbleChartType = "bubble"
// 	PieChartType = "pie"
// 	ScatterChartType = "scatter"
// )

// // TimeUnit enums
// const(
// 	MillisecondTimeUnit TimeUnit = "millisecond"
// 	SecondTimeUnit = "second"
// 	MinuteTimeUnit = "minute"
// 	HourTimeUnit = "hour"
// 	DayTimeUnit = "day"
// 	WeekTimeUnit = "week"
// 	MonthTimeUnit = "month"
// 	QuarterTimeUnit = "quarter"
// 	YearTimeUnit = "year"
// )

// // ScaleType enums
// const(
// 	CategoryScaleType ScaleType = "category"
// 	LinearScaleType = "linear"
// 	LogarithmicScaleUnit = "logarithmic"
// 	TimeScaleUnit = "time"
// 	RadialLinearScaleUnit = "radialLinear"
// )

// // PointStyle enums
// const(
// 	CirclePointStyle PointStyle = "circle"
// 	CrossPointStyle = "cross"
// 	CrossRotPointStyle = "crossRot"
// 	DashPointStyle = "dash"
// 	LinePointStyle = "line"
// 	RectPointStyle = "rect"
// 	RectRoundedPointStyle = "rectRounded"
// 	RectRotPointStyle = "rectPot"
// 	StarPointStyle = "star"
// 	TrianglePointStyle = "triangle"
// )

// // PositionType enums
// const(
// 	LeftPositionType PositionType = "left"
// 	RightPositionType = "right"
// 	TopPositionType = "top"
// 	BottomPositionType = "bottom"
// 	ChartAreaPositionType = "chartArea"
// )

// type ChartType string
// type TimeUnit string
// type ScaleType string
// type PointStyle string
// type PositionType string

type ChartData struct {
	Labels   []string        `json:"labels"`
	Datasets []ChartDataSets `json:"datasets"`
}

type ChartDataSets struct {
	BackgroundColor           []string      `json:"backgroundColor"`
	BorderColor               []string      `json:"borderColor"`
	Data                      []interface{} `json:"data"`
	Fill                      string        `json:"fill"`
	HoverBackgroundColor      []string      `json:"hoverBackgroundColor"`
	HoverBorderColor          []string      `json:"hoverBorderColor"`
	Label                     string        `json:"label"`
	PointBorderColor          []string      `json:"pointBorderColor"`
	PointBackgroundColor      []string      `json:"pointBackgroundColor"`
	PointHoverBackgroundColor []string      `json:"pointHoverBackgroundColor"`
	PointHoverBorderColor     []string      `json:"pointHoverBorderColor"`
}

// // ChartConfig represents
// type ChartConfig struct {
// 	Type string      `json:"type"`
// 	Data []ChartData `json:"data"`
// }

// // ChartData represents datasets with labels for the datasets
// type ChartData struct {
// 	Labels   []string      `json:"labels"`
// 	Datasets []interface{} `json:"datasets"`
// }

// // PieProperties are properties used in pie chart dataset
// type PieProperties struct {
// 	BackgroundColor      string    `json:"backgroundColor"`
// 	BorderAlign          string    `json:"borderAlign"`
// 	BorderColor          string    `json:"borderColor"`
// 	BorderWidth          float64   `json:"borderWidth"`
// 	Data                 []float64 `json:"data"`
// 	HoverBackgroundColor string    `json:"hoverBackgroundColor"`
// 	HoverBorderColor     string    `json:"hoverBorderColor"`
// 	HoverBorderWidth     float64   `json:"hoverBorderWidth"`
// 	Weight               float64   `json:"weight"`
// }

// // PieOptions are options used in pie chart dataset
// type PieOptions struct {
// 	CutoutPercentage float64 `json:"cutoutPercentage"`
// 	Rotation         float64 `json:"rotation"`
// 	Circumference    float64 `json:"circumference"`
// 	AnimateRotate    float64 `json:"animation.animateRotate"`
// 	AnimateScale     float64 `json:"animation.animateScale"`
// }
