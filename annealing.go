// Copyright ©2017 The ezgliding Authors.
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package igc

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// NewSimAnnealingOptimizer ...
func NewSimAnnealingOptimizer() Optimizer {
	return NewSimAnnealingOptimizerParams(1000, 1, 0.003, time.Now().UTC().UnixNano())
}

// NewSimAnnealingOptimizerParams returns a BruteForceOptimizer with the given characteristics.
func NewSimAnnealingOptimizerParams(startTemperature float64, minTemperature float64,
	alpha float64, seed int64) Optimizer {
	rand.Seed(seed)
	return &simAnnealingOptimizer{
		currentTemperature: startTemperature,
		minTemperature:     minTemperature,
		alpha:              alpha,
	}
}

type simAnnealingOptimizer struct {
	score              Score
	currentTemperature float64
	minTemperature     float64
	alpha              float64
	track              *Track
	nPoints            int
	candidate          Candidate
	best               Candidate
}

func (sa *simAnnealingOptimizer) initialize(track *Track, nPoints int, score Score) {
	sa.track = track
	sa.nPoints = nPoints
	sa.score = score
	sa.candidate = NewCandidateRandom(nPoints, track)
	sa.best = Candidate(sa.candidate)
}

func (sa *simAnnealingOptimizer) neighbour() Candidate {
	return sa.candidate.Neighbour()
}

func (sa *simAnnealingOptimizer) acceptanceProb(task Task) float64 {
	diff := sa.score(sa.candidate.Task) - sa.score(task)
	if diff > 0 {
		return 1.0
	}
	return math.E * (diff / sa.currentTemperature)
}

type RunLog struct {
	Candidates  []Candidate
	Bests       []Candidate
	Temperature []float64
}

func (sa *simAnnealingOptimizer) Optimize(track Track, nPoints int, score Score) (Task, error) {
	var acceptanceProb float64
	var candidate Candidate

	sa.initialize(&track, nPoints, score)

	log := RunLog{Candidates: make([]Candidate, 1), Bests: make([]Candidate, 1), Temperature: make([]float64, 1)}
	// loop while the temperature is above min
	for sa.currentTemperature > sa.minTemperature {
		candidate = sa.neighbour()
		acceptanceProb = sa.acceptanceProb(candidate.Task)
		if acceptanceProb > rand.Float64() {
			sa.candidate = candidate
		}
		if sa.score(candidate.Task) > sa.score(sa.best.Task) {
			sa.best = Candidate(candidate)
		}
		sa.currentTemperature *= (1 - sa.alpha)
		log.Candidates = append(log.Candidates, sa.candidate)
		log.Bests = append(log.Bests, sa.best)
		log.Temperature = append(log.Temperature, sa.currentTemperature)
	}
	txt, _ := json.MarshalIndent(log, "", "  ")
	fmt.Printf("%v\n", string(txt))
	return sa.best.Task, nil
}