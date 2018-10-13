package Predict

import (
	. "MidtermForecast/APIs"
	. "MidtermForecast/Utils"
	"encoding/csv"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/mathext"
	"gonum.org/v1/gonum/optimize"
	"gonum.org/v1/gonum/stat/distmv"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var Senate2016results = map[string][2]float64{
	"AL": {748709, 1335104},
	"AK": {36200, 138149},
	"AZ": {1031245, 1359267},
	"AR": {400602, 661984},
	"CA": {7542753, 4701417},
	"CO": {1370710, 1215318},
	"CT": {1008714, 552621},
	"FL": {4122088, 4835191},
	"GA": {1599726, 2135806},
	"HI": {306604, 92653},
	"ID": {188249, 449017},
	"IL": {3012940, 2184692},
	"IN": {1158947, 1423991},
	"IA": {549460, 926007},
	"KS": {379740, 732376},
	"KY": {813246, 1090177},
	"LA": {578750, 984625},
	"MD": {1659907, 972557},
	"MO": {1300200, 1378458},
	"NV": {521994, 495079},
	"NH": {354649, 353632},
	"NY": {5221945, 2009335},
	"NC": {2128165, 2395376},
	"ND": {58116, 268788},
	"OH": {1996908, 3118567},
	"OK": {355911, 980892},
	"OR": {1105119, 651106},
	"PA": {2865012, 2951702},
	"SC": {757022, 1241609},
	"SD": {104125, 265494},
	"UT": {301860, 760241},
	"VT": {192243, 103637},
	"WA": {745421, 381004},
	"WI": {1380335, 1479471},
}

var generic_ballot_map = map[int]float64{
	1998: 0, // TODO (not used though)
	2000: 0, // TODO (not used though)
	2002: 0, // TODO (not used though)
	2004: 0, // TODO (not used though)
	2006: 0, // TODO (not used though)
	2008: float64(65237840-52249491) / float64(65237840+52249491),
	2010: float64(38980192-44827441) / float64(38980192+44827441),
	2012: float64(59645531-58228253) / float64(59645531+58228253),
	2014: float64(35624357-40081282) / float64(35624357+40081282),
	2016: float64(61776554-63173815) / float64(61776554+63173815),
}

func LoadHistoricalData() (polls []map[string][]Poll, results []map[string][2]float64, electionDates []time.Time, experts_ratings [][]map[string][2]float64, incumbents []map[string]string, expert_pvis [][]map[string]float64, fundraising []map[string]float64, elasticities []map[string]float64) {
	records := LoadCache("https://raw.githubusercontent.com/fivethirtyeight/data/master/pollster-ratings/raw-polls.csv", "cache/raw_polls.csv", -1, func(r io.Reader) interface{} {
		records, err := csv.NewReader(r).ReadAll()
		if err != nil {
			panic(err)
		}
		return records
	}).([][]string)
	generals := map[string]time.Time{
		"1998": time.Date(1998, time.November, 3, 0, 0, 0, 0, time.UTC),
		"2000": time.Date(2000, time.November, 7, 0, 0, 0, 0, time.UTC),
		"2002": time.Date(2002, time.November, 5, 0, 0, 0, 0, time.UTC),
		"2004": time.Date(2004, time.November, 2, 0, 0, 0, 0, time.UTC),
		"2006": time.Date(2006, time.November, 7, 0, 0, 0, 0, time.UTC),
		"2008": time.Date(2008, time.November, 4, 0, 0, 0, 0, time.UTC),
		"2010": time.Date(2010, time.November, 2, 0, 0, 0, 0, time.UTC),
		"2012": time.Date(2012, time.November, 6, 0, 0, 0, 0, time.UTC),
		"2014": time.Date(2014, time.November, 4, 0, 0, 0, 0, time.UTC),
		"2016": time.Date(2016, time.November, 8, 0, 0, 0, 0, time.UTC),
	}
	yearidx := map[string]int{}
	electionDates = []time.Time{}
	experts_ratings = [][]map[string][2]float64{}
	incumbents = []map[string]string{}
	expert_pvis = [][]map[string]float64{}
	fundraising = []map[string]float64{}
	elasticities = []map[string]float64{}
	// Only do Senate/House/Gov 2012/2014/2016 for now because those are all I can get proper historical data for
	for _, year := range []string{ /*"1998", "2000", "2002", "2004", "2006", "2008", "2010", */ "2012", "2014", "2016"} {
		date := generals[year]
		for _, r := range []string{"Sen-G", "House-G", "Gov-G" /*, "Pres-G"*/} {
			if r == "Gov-G" && year == "2008" {
				continue // No way to get results
			}
			yearidx[year+" "+r] = len(electionDates)
			electionDates = append(electionDates, date)

			// Cook
			var crats map[string][2]float64
			var cincs map[string]string
			var cpvis map[string]float64
			f := true
			if r == "House-G" {
				if year == "2008" {
					crats, cincs, cpvis = LoadCookHouseRatingsHtml(139081) // Will give wrong PVIs because Cook gives PVIs of current districts even for historical
				} else if year == "2012" {
					crats, cincs, cpvis = LoadCookHouseRatingsHtml(139119)
				} else if year == "2014" {
					crats, cincs, cpvis = LoadCookHouseRatingsHtml(139258)
				} else if year == "2016" {
					crats, cincs, cpvis = LoadCookHouseRatingsHtml(139361)
				} else {
					f = false
				}
				fundraising = append(fundraising, LoadFECRaces("H", generals[year].Year()))
				elasticities = append(elasticities, Load538HouseElasticities())
			} else if r == "Sen-G" {
				if year == "2008" {
					crats, cincs, cpvis = LoadCookSenateRatingsHtml(139080) // Will give wrong PVIs because Cook gives PVIs of current districts even for historical
				} else if year == "2012" {
					crats, cincs, cpvis = LoadCookSenateRatingsHtml(139117)
				} else if year == "2014" {
					crats, cincs, cpvis = LoadCookSenateRatingsHtml(139256)
				} else if year == "2016" {
					crats, cincs, cpvis = LoadCookSenateRatingsHtml(139360)
				} else {
					f = false
				}
				fundraising = append(fundraising, LoadFECRaces("S", generals[year].Year()))
				elasticities = append(elasticities, Load538SenateElasticities())
			} else if r == "Gov-G" {
				if year == "2008" {
					crats, cincs, cpvis = LoadCookGovRatingsHtml(139082) // Will give wrong PVIs because Cook gives PVIs of current districts even for historical
				} else if year == "2012" {
					crats, cincs, cpvis = LoadCookGovRatingsHtml(139101)
				} else if year == "2014" {
					crats, cincs, cpvis = LoadCookGovRatingsHtml(139257)
				} else if year == "2016" {
					crats, cincs, cpvis = LoadCookGovRatingsHtml(139343)
				} else {
					f = false
				}
				fundraising = append(fundraising, nil)
				elasticities = append(elasticities, Load538SenateElasticities())
			} else {
				f = false
				fundraising = append(fundraising, nil)
				elasticities = append(elasticities, Load538SenateElasticities())
			}
			if f {
				experts_ratings = append(experts_ratings, []map[string][2]float64{crats})
				incumbents = append(incumbents, cincs)
				expert_pvis = append(expert_pvis, []map[string]float64{cpvis})
			} else {
				experts_ratings = append(experts_ratings, nil)
				incumbents = append(incumbents, nil)
				expert_pvis = append(expert_pvis, nil)
			}
		}
	}
	polls = make([]map[string][]Poll, len(yearidx))
	results = make([]map[string][2]float64, len(yearidx))
	var n int
	for _, record := range records {
		if record[10] != "Democrat" || record[12] != "Republican" || record[9] == "" {
			continue
		}
		year := record[2] + " " + record[4]
		if _, ok := yearidx[year]; !ok {
			continue
		}
		st := record[3]
		if st == "US" || st == "USA" {
			continue
		}
		if strings.Contains(st, "-") {
			if len(st) == 4 {
				st = st[:3] + "0" + st[3:]
			}
		}
		if record[5] == "Sen-GS" { // Special
			st += "-2"
		}
		if st == "HI-2" {
			st = "HI"
		}
		if st == "AK-01" || st == "DE-01" || st == "MT-01" || st == "ND-01" || st == "SD-01" || st == "VT-01" || st == "WY-01" {
			st = st[:3] + "00"
		}
		var poll Poll
		poll.Pollster = record[6]
		start, err := time.Parse("1/2/06", record[8])
		if err != nil {
			panic(err)
		}
		poll.StartDate = start
		poll.EndDate = start
		poll.Subpopulation = "LV"
		num, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			panic(err)
		}
		poll.Number = num
		dpct, err := strconv.ParseFloat(record[11], 64)
		if err != nil {
			panic(err)
		}
		rpct, err := strconv.ParseFloat(record[13], 64)
		if err != nil {
			panic(err)
		}
		dact, err := strconv.ParseFloat(record[17], 64)
		if err != nil {
			panic(err)
		}
		ract, err := strconv.ParseFloat(record[18], 64)
		if err != nil {
			panic(err)
		}
		poll.Candidates = map[string]float64{"D": dpct / 100.0, "R": rpct / 100.0}
		if polls[yearidx[year]] == nil {
			polls[yearidx[year]] = map[string][]Poll{}
			results[yearidx[year]] = map[string][2]float64{}
		}
		if _, ok := polls[yearidx[year]][st]; !ok {
			polls[yearidx[year]][st] = make([]Poll, 0)
		}
		polls[yearidx[year]][st] = append(polls[yearidx[year]][st], poll)
		results[yearidx[year]][st] = [2]float64{dact, ract}
		n++
	}
	for year := range yearidx {
		if _, ok := map[string]struct{}{"2012 House-G": struct{}{}, "2014 House-G": struct{}{}, "2016 House-G": struct{}{}}[year]; ok {
			for st, r := range LoadCNNHouseResults(year[:4]) {
				results[yearidx[year]][st] = r
			}
		} else if _, ok := map[string]struct{}{"2012 Sen-G": struct{}{}, "2014 Sen-G": struct{}{}, "2016 Sen-G": struct{}{}}[year]; ok {
			for st, r := range LoadCNNSenateResults(year[:4]) {
				results[yearidx[year]][st] = r
			}
		} else if _, ok := map[string]struct{}{"2012 Gov-G": struct{}{}, "2014 Gov-G": struct{}{}, "2016 Gov-G": struct{}{}}[year]; ok {
			for st, r := range LoadCNNGovResults(year[:4]) {
				results[yearidx[year]][st] = r
			}
		}
	}
	if idx, ok := yearidx["2008 Sen-G"]; ok {
		results[idx] = LoadNYTSenateResults("2008")
	}
	if idx, ok := yearidx["2008 House-G"]; ok {
		results[idx] = LoadNYTHouseResults("2008")
	}
	if idx, ok := yearidx["2016 Sen-G"]; ok {
		polls[idx] = Load2016SenatePolls()
		results[idx] = Senate2016results
	}
	return polls, results, electionDates, experts_ratings, incumbents, expert_pvis, fundraising, elasticities
}

func linfit(X, Y []float64) (a, b, rmse float64) {
	var SY, SY2, SXY, SX, SX2, n float64
	for i := 0; i < len(Y); i++ {
		SY += Y[i]
		SY2 += Y[i] * Y[i]
		SXY += X[i] * Y[i]
		SX += X[i]
		SX2 += X[i] * X[i]
		n += 1
	}
	a = (SY*SX2 - SX*SXY) / (n*SX2 - SX*SX)
	b = (n*SXY - SX*SY) / (n*SX2 - SX*SX)
	var e float64
	for i := 0; i < len(Y); i++ {
		e += math.Pow((a+b*X[i])-Y[i], 2)
	}
	rmse = math.Sqrt(e / n)
	return
}

func getPollingErr(polls []Poll, result [2]float64, electionDate, now time.Time) (float64, int) {
	r := result[0] / (result[0] + result[1])
	ppolls := make([]Poll, 0, len(polls))
	for _, poll := range polls {
		if poll.EndDate.Before(now) {
			ppolls = append(ppolls, poll)
		}
	}
	params := GetConcentrationParams(ppolls, map[string]string{"D": "D", "R": "R"}, electionDate)
	s := params["D"] / (params["D"] + params["R"])
	return r - s, len(ppolls)
}

func OptimizePollShift() {
	stpolls, results, dates, _, _, _, _, _ := LoadHistoricalData()
	errs := make([]float64, 0)
	days := make([]float64, 0)
	for idx, stpoll := range stpolls {
		for st, polls := range stpoll {
			for _, poll := range polls {
				e, _ := getPollingErr([]Poll{poll}, results[idx][st], dates[idx], dates[idx])
				err := math.Abs(e)
				//d := poll.EndDate.Sub(poll.StartDate).Hours() / 24.0
				ds := dates[idx].Sub(poll.EndDate).Hours() / 24.0
				if ds < 0 {
					continue
				}
				errs = append(errs, err)
				days = append(days, math.Sqrt(ds))
			}
		}
	}
	a, b, rmse := linfit(days, errs)
	fmt.Println("Poll shift:", a, b, rmse)
	RAND_POLL_SHIFT = a
	DAILY_POLL_SHIFT = b
}

func OptimizeNationalShift() {
	stpolls, results, dates, _, _, _, _, _ := LoadHistoricalData()

	ntnldays := make([]float64, 0)
	ntnlerrs := make([]float64, 0)

	// TODO: This needs to be improved/more accurate - separate out the national fit from the race fit
	// TODO: Optimize this separately for Sen/Gov and House
	for idx, stpoll := range stpolls {
		for days := 400; days >= 0; days-- {
			ds := math.Sqrt(float64(days))
			var ntnlerr, ntnlerr2, ntnln float64
			for st, polls := range stpoll {
				e, n := getPollingErr(polls, results[idx][st], dates[idx].Add(-24*time.Hour*time.Duration(days)), dates[idx])
				if n == 0 {
					continue
				}
				ntnlerr += e
				ntnlerr2 += e * e
				ntnln += 1
			}
			if ntnln <= 3 {
				continue
			}
			err := math.Sqrt(ntnlerr2*ntnln-ntnlerr*ntnlerr) / (ntnln - 1)
			ntnlerrs = append(ntnlerrs, err)
			ntnldays = append(ntnldays, ds)
		}
	}
	a, b, rmse := linfit(ntnldays, ntnlerrs)
	fmt.Println("National shift:", a, b, rmse)
	RAND_NATIONAL_SHIFT = a
	DAILY_NATIONAL_SHIFT = b
}

func OptimizeRaceShift() {
	stpolls, results, dates, _, _, _, _, _ := LoadHistoricalData()

	racedays := make([]float64, 0)
	raceerrs := make([]float64, 0)

	// TODO: This needs to be improved/more accurate - separate out the national fit from the race fit
	// TODO: Optimize this separately for Sen/Gov and House
	for idx, stpoll := range stpolls {
		for days := 400; days >= 0; days-- {
			ds := math.Sqrt(float64(days))
			for st, polls := range stpoll {
				e, n := getPollingErr(polls, results[idx][st], dates[idx].Add(-24*time.Hour*time.Duration(days)), dates[idx])
				if n == 0 {
					continue
				}
				raceerrs = append(raceerrs, math.Abs(e))
				racedays = append(racedays, ds)
			}
		}
	}
	a, b, rmse := linfit(racedays, raceerrs)
	fmt.Println("Race shift:", a, b, rmse)
	RAND_RACE_SHIFT = a
	DAILY_RACE_SHIFT = b
}

// Minimizes sum(pow(x[i][j]*b[j]-y[i], 2) for j in x[i] for i in x) for b[j]
func lsfit(X [][]float64, Y []float64) (B []float64, rmse float64) {
	Xm := mat.NewDense(len(X), len(X[0]), nil)
	for i, Xr := range X {
		Xm.SetRow(i, Xr)
	}
	Yv := mat.NewVecDense(len(Y), Y)
	Bv := mat.NewVecDense(len(X[0]), nil)
	err := Bv.SolveVec(Xm, Yv)
	if err != nil {
		panic(err)
	}
	for i, Xr := range X {
		var xr float64
		for j, x := range Xr {
			xr += x * Bv.At(j, 0)
		}
		rmse += (xr - Y[i]) * (xr - Y[i])
	}
	rmse /= float64(len(X))
	rmse = math.Sqrt(rmse)
	return Bv.RawVector().Data, rmse
}

func OptimizePVIs() {
	_, results, dates, _, incumbents, expert_pvis, fundraising, elasticities := LoadHistoricalData()
	// Linear fit: cook_pvi*COOK_PVI_WEIGHT+past_pvi*PAST_PVI_WEIGHT+log_fundraising_ratio*FUNDRAISING_MULTIPLIER+overall_pvi*OVERALL_PVI_WEIGHT+incumbent_advantange*INCUMBENT_ADVANTAGE_PVI = d_margin
	var rmse float64
	for t := 0; t < 10; t++ {
		// Do this iteratively to learn real INCUMBENT_ADVANTAGE_PVI
		X := make([][]float64, 0)
		Y := make([]float64, 0)
		for idx, res := range results {
			var cook_pvis = expert_pvis[idx][0]
			var past_pvis = map[string]float64{}
			if len(res) > 200 {
				// Only have good past results data for House races
				var pidx = idx - 1
				for len(results[pidx]) < 200 && pidx > 0 {
					pidx--
				}
				if pidx != 0 || len(results[pidx]) > 200 {
					for pk, presult := range results[pidx] {
						past_pvis[pk] = (presult[0]-presult[1])/(presult[0]+presult[1]) - generic_ballot_map[dates[pidx].Year()]
						if incumbents[pidx][pk] == "R" {
							past_pvis[pk] += INCUMBENT_ADVANTAGE_PVI
						} else if incumbents[pidx][pk] == "D" {
							past_pvis[pk] -= INCUMBENT_ADVANTAGE_PVI
						}
					}
				}
			}
			var log_fundraising_ratios = map[string]float64{}
			for k, v := range fundraising[idx] {
				log_fundraising_ratios[k] = math.Log(v) / math.Ln2
			}
			generic_ballot := generic_ballot_map[dates[idx].Year()]
			var incumbent = map[string]float64{}
			for k, i := range incumbents[idx] {
				if i == "D" {
					incumbent[k] = 1.0
				} else if i == "R" {
					incumbent[k] = -1.0
				}
			}
			for k, r := range res {
				if past_pvis[k] == 0 {
					continue
				}
				if r[0] == 0 || r[1] == 0 {
					continue
				}
				elasticity := elasticities[idx][k]
				if elasticity == 0.0 {
					elasticity = 1.0
				}
				X = append(X, []float64{cook_pvis[k], past_pvis[k], log_fundraising_ratios[k], generic_ballot * elasticity, incumbent[k]})
				Y = append(Y, (r[0]-r[1])/(r[0]+r[1]))
			}
		}
		var B []float64
		B, rmse = lsfit(X, Y)
		COOK_PVI_WEIGHT, PAST_PVI_WEIGHT, FUNDRAISING_MULTIPLIER, OVERALL_PVI_WEIGHT, INCUMBENT_ADVANTAGE_PVI = B[0], B[1], B[2], B[3], B[4]
	}
	fmt.Println("Cook PVI weight:", COOK_PVI_WEIGHT, "Past PVI weight:", PAST_PVI_WEIGHT, "Fundraising multiplier:", FUNDRAISING_MULTIPLIER, "Overall PVI weight:", OVERALL_PVI_WEIGHT, "Incumbent Advantage PVI:", INCUMBENT_ADVANTAGE_PVI, "RMSE:", rmse)
}

func addFloat64(val *float64, delta float64) {
	for {
		old := *val
		new := old + delta
		if atomic.CompareAndSwapUint64(
			(*uint64)(unsafe.Pointer(val)),
			math.Float64bits(old),
			math.Float64bits(new),
		) {
			break
		}
	}
}

func OptimizeSources() {
	stpolls, results, dates, expert_ratings, incumbents, expert_pvis, fundraising, elasticities := LoadHistoricalData()
	prob := optimize.Problem{
		Func: func(vals []float64) float64 {
			POLLING_WEIGHT = vals[0]
			FUNDAMENTALS_WEIGHT = vals[1]
			COOK_WEIGHT = vals[2]
			if POLLING_WEIGHT < 0 || FUNDAMENTALS_WEIGHT < 0 || COOK_WEIGHT < 0 {
				return math.NaN() // Minimize doesn't like -Inf
			}
			var llscore float64
			var wg sync.WaitGroup
			for days := 400; days >= 0; days-- {
				wg.Add(1)
				go func(days int) {
					defer wg.Done()
					for idx, rs := range results {
						generic_ballot := generic_ballot_map[dates[idx].Year()]
						past_pvis := make(map[string]float64)
						if len(rs) > 200 {
							// House
							var pidx = idx - 1
							for len(results[pidx]) < 200 && pidx > 0 {
								pidx--
							}
							if pidx != 0 || len(results[pidx]) > 200 {
								for pk, presult := range results[pidx] {
									past_pvis[pk] = (presult[0]-presult[1])/(presult[0]+presult[1]) - generic_ballot_map[dates[pidx].Year()]
									if incumbents[pidx][pk] == "R" {
										past_pvis[pk] += INCUMBENT_ADVANTAGE_PVI
									} else if incumbents[pidx][pk] == "D" {
										past_pvis[pk] -= INCUMBENT_ADVANTAGE_PVI
									}
								}
							}
						}
						stppolls := make(map[string][]Poll, len(stpolls[idx]))
						for st, polls := range stpolls[idx] {
							ppolls := make([]Poll, 0, len(polls))
							for _, poll := range polls {
								if poll.EndDate.Before(dates[idx].Add(-time.Duration(days) * 24 * time.Hour)) {
									ppolls = append(ppolls, poll)
								}
							}
							if len(ppolls) > 0 {
								stppolls[st] = ppolls
							}
						}
						if len(stppolls) == 0 {
							continue
						}
						polling_ratings := GetPollingRatings(stppolls, dates[idx])
						pvi_ests := make([]map[string]float64, 0)
						pvi_ests = append(pvi_ests, expert_pvis[idx]...)
						pvi_weights := []float64{COOK_PVI_WEIGHT}
						if len(past_pvis) > 0 {
							pvi_ests = append(pvi_ests, past_pvis)
							pvi_weights = append(pvi_weights, PAST_PVI_WEIGHT)
						} else {
							continue
						}

						fundamentals_ratings, _ := GetFundamentalsRatings(incumbents[idx], fundraising[idx], pvi_ests, pvi_weights, nil, elasticities[idx], generic_ballot)
						experts_ratings := CombineRatings([]map[string][2]float64{expert_ratings[idx][0]}, []float64{COOK_WEIGHT})
						ratings := MergeRatings([]map[string][2]float64{polling_ratings, fundamentals_ratings, experts_ratings})
						for st, rat := range ratings {
							if rat[0] < 0 || rat[1] < 0 {
								llscore = math.NaN()
								return
							}
							//rat[0], rat[1] = AdjustRaceError(rat[0]+1, rat[1]+1, float64(days))
							if rs[st][0] > rs[st][1] {
								addFloat64(&llscore, math.Log(mathext.RegIncBeta(rat[1], rat[0], 0.5)))
							} else if rs[st][0] < rs[st][1] {
								addFloat64(&llscore, math.Log(mathext.RegIncBeta(rat[0], rat[1], 0.5)))
							}
						}
					}
				}(days)
			}
			wg.Wait()
			return -llscore
		},
	}
	//vals := []float64{0.749368253464103, 25.153525945790587, 301.1128810616409} // Precomputed values
	vals := []float64{0.8102851928864727, 29.45179266376384, 297.88292721763725} // Precomputed values
	if false {
		// Brute-force guess-and-check
		result, err := optimize.Minimize(prob, vals, &optimize.Settings{FuncEvaluations: 1000}, &optimize.GuessAndCheck{Rander: distmv.NewUniform([]distmv.Bound{{0.0, 10000.0}, {0.0, 10000.0}, {0.0, 10000.0}}, nil)})
		if err != nil {
			panic(err)
		}
		fmt.Println("Guess-And-Check Sources:", result)
		vals = result.X
	}
	if true {
		// Nedler-Mead optimize vals
		result, err := optimize.Minimize(prob, vals, &optimize.Settings{FuncEvaluations: 1000}, &optimize.NelderMead{})
		if err != nil {
			panic(err)
		}
		fmt.Println("Nelder-Mead Sources:", result)
		vals = result.X
	}
	POLLING_WEIGHT = vals[0]
	FUNDAMENTALS_WEIGHT = vals[1]
	COOK_WEIGHT = vals[2]
	fmt.Println("Polling weight:", vals[0], "Fundamentals weight:", vals[1], "Cook weight:", vals[2])
}

func OptimizeHistorical() {
	OptimizePollShift()
	OptimizeNationalShift()
	OptimizeRaceShift()
	OptimizePVIs()
	OptimizeSources()

	/*POLLING_WEIGHT = 8.0
	COOK_PVI_WEIGHT = 2.0
	INCUMBENT_ADVANTAGE_PVI = 0.07
	OVERALL_PVI_WEIGHT = 3.0
	FUNDAMENTALS_WEIGHT = 230.0
	COOK_WEIGHT = 1000.0

	RAND_POLL_SHIFT = 0.02
	DAILY_POLL_SHIFT = 0.0025
	RAND_RACE_SHIFT = 0.04
	DAILY_RACE_SHIFT = 0.0
	RAND_NATIONAL_SHIFT = 0.03
	DAILY_NATIONAL_SHIFT = 0.0*/
}
