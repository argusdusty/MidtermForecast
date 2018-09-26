package Predict

import (
	"gonum.org/v1/gonum/mathext"
	"math"
	"runtime"
	"sync"
)

var (
	// Pre-computed in historical_optimize.go
	RAND_NATIONAL_SHIFT  = 0.036811891821249415   // Base national error
	DAILY_NATIONAL_SHIFT = 1.0850935968504596e-05 // Per-day average national drift
	RAND_RACE_SHIFT      = 0.028265923390389497   // Base per-race error
	DAILY_RACE_SHIFT     = 1.2133332224145488e-05 // Per-day per-race average drift

	N = 1 << 22 // Number of race simulations. Can go lower without much loss of accuracy, or higher if you want.
)

func Prob(races map[string][2]float64, days, rw float64) ([]float64, map[string]RaceProbability) {
	num_seats := len(races) + 1
	r := make([]float64, num_seats)
	var wg sync.WaitGroup
	M := N / len(races)
	var lock sync.Mutex
	race_probs := map[string]RaceProbability{}
	worker_chan := make(chan int, N)
	for n := 0; n < runtime.NumCPU(); n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			partial_race_probs := map[string][3]float64{}
			c := make([]float64, num_seats)
			t := make([]float64, num_seats)
			for i := range worker_chan {
				for k := range partial_race_probs {
					partial_race_probs[k] = [3]float64{}
				}
				for j, _ := range c {
					c[j] = 0
					t[j] = 0
				}
				c[0] = 1
				// Inverse of normal CDF with mean of 0 and standard deviation of A+B*sqrt(days). Used to do a numeric integration that rapidly converges.
				shift := math.Sqrt2 * (RAND_NATIONAL_SHIFT + (DAILY_NATIONAL_SHIFT * math.Sqrt(days))) * rw * math.Erfinv(2*(float64(i)+0.5)/float64(M)-1)
				for st, rating := range races {
					for j, _ := range t {
						t[j] = 0
					}
					var pb, ipb float64
					alpha := rating[0]
					beta := rating[1]
					if alpha == 0.0 {
						pb = 0.0
						ipb = 1.0
					} else if beta == 0.0 {
						pb = 1.0
						ipb = 0.0
					} else {
						if shift < 0 {
							alpha, beta = alpha+alpha*shift, beta-alpha*shift
							if shift < -1 {
								alpha = 0.0
								beta = 1.0
							}
						} else {
							alpha, beta = alpha+beta*shift, beta-beta*shift
							if shift > 1 {
								alpha = 1.0
								beta = 0.0
							}
						}
						pb = mathext.RegIncBeta(beta, alpha, 0.5)
						ipb = mathext.RegIncBeta(alpha, beta, 0.5)
					}
					rp := partial_race_probs[st]
					partial_race_probs[st] = [3]float64{rp[0] + pb, rp[1] + alpha, rp[2] + beta}

					for k, x := range c {
						if x == 0 {
							continue
						}
						t[k] += x * ipb
						t[k+1] += x * pb
					}
					copy(c, t)
				}
				lock.Lock()
				for j, _ := range c {
					r[j] += c[j] / float64(M)
				}
				for st, p := range partial_race_probs {
					rp := race_probs[st]
					rp.Race = st
					rp.DemWinProbability += p[0] / float64(M)
					if rp.ConcentrationParams == nil {
						rp.ConcentrationParams = map[string]float64{"D": 0.0, "R": 0.0}
					}
					rp.ConcentrationParams["D"] += p[1] / float64(M)
					rp.ConcentrationParams["R"] += p[2] / float64(M)
					race_probs[st] = rp
				}
				lock.Unlock()
			}
		}()
	}
	for i := 0; i < M; i++ {
		worker_chan <- i
	}
	close(worker_chan)
	wg.Wait()
	return r, race_probs
}

func AdjustRaceError(alpha, beta, days float64) (a, b float64) {
	if alpha == 0.0 || beta == 0.0 {
		return alpha, beta
	}
	berr := math.Sqrt(alpha * beta / ((alpha + beta) * (alpha + beta) * (alpha + beta + 1)))
	err := berr + math.Abs(RAND_RACE_SHIFT+DAILY_RACE_SHIFT*math.Sqrt(days))
	//w := (alpha*beta)/(err*err*(alpha+beta)*(alpha+beta)*(alpha+beta)) - 1.0/(alpha+beta) // - can go negative, bad
	w := math.Pow(berr/err, 2)
	a = alpha * w
	b = beta * w
	return
}
