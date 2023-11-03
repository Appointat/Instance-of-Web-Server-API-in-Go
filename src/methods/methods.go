package methods

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

/*
Types of basic methods

Les ints seront représentées par des entiers.
Les profils de préférences sont telles que si profile est un profil, profile[12] représentera les préférences du votant 12. Les ints sont classée de la préférée à la moins préférée :  profile[12][0] represente l'int préférée du votant 12.
Enfin, les méthodes de vote renvoient un décompte sous forme d'une map qui associe à chaque int un entier : plus cet entier est élevé, plus l'int a de points et plus elle est préférée pour le groupe compte tenu de la méthode considérée.
*/
type Profile [][]int
type Count map[int]int // map associant à chaque int un entier

// e.g. of type
// profile[12][0] is the first choice of voter 12
// profile[12][1] is the second choice of voter 12
// Count is a map associating each int with an integer
// e.g. count[12] = 3 means that int 12 has 3 votes

// return the index of the alt in the prefs, or -1 if not found
func rank(alt int, prefs []int) int {
	for i, a := range prefs {
		if a == alt {
			return i
		}
	}
	return -1
}

// return true if alt1 is preferred to alt2 in the profile
func isPref(alt1, alt2 int, prefs []int) bool {
	//we need to consider the situation where alt1 or alt2 is not in the preference list
	rank1 := rank(alt1, prefs)
	rank2 := rank(alt2, prefs)
	//if alt1 is not in the preference list, it is not preferred to alt2
	//if alt2 is not in the preference list, it is preferred to alt1
	//if both are not in the preference list, return false
	if (rank1 == -1) && (rank2 == -1) {
		return false
	}
	if rank1 == -1 {
		return false
	}
	if rank2 == -1 {
		return true
	}
	return rank1 < rank2
}

// return the best ints for the given profile in order of preference
func maxCount(count Count) (bestAlts []int) {
	max := 0
	for alt, cnt := range count {
		if cnt > max {
			max = cnt
			bestAlts = []int{alt}
			// TODO: maxCount函数应该根据test函数应返回一个map，但是题目要求却返回一个数组
		} else if cnt == max {
			bestAlts = append(bestAlts, alt)
		}
	}
	return bestAlts

}

// check if the given profile, e.g. that they are all complete and that each alt only appears once per pref
func checkProfile(prefs Profile) error {
	// Check if the profile is complete
	if len(prefs) == 0 {
		return errors.New("empty profile")
	}

	numAlts := len(prefs[0])
	// Check the length of the preference list
	for i, pref := range prefs {
		if len(pref) != numAlts {
			return fmt.Errorf("preference list %d is incomplete", i)
		}

		seen := make(map[int]bool)
		for _, alt := range pref {
			if seen[alt] {
				return fmt.Errorf("int %d appears more than once in preference list %d", alt, i)
			}
			seen[alt] = true
		}

		// Ensure that each int is presented in the preference list
		for _, pref := range prefs {
			for _, alt := range pref {
				if !seen[alt] {
					return fmt.Errorf("int %d is missing in preference list %d", alt, i)
				}
			}
		}
	}

	return nil
}

// check if the given profile, e.g. that they are all complete and that each alt only appears once per pref
func checkProfileint(prefs Profile, alts []int) error {
	altMap := make(map[int]bool)
	for _, alt := range alts {
		altMap[alt] = true
	}

	for i, pref := range prefs {
		// check the length of the preference list
		if len(pref) != len(alts) {
			return fmt.Errorf("preference list %d is incomplete", i)
		}

		seen := make(map[int]bool)
		for _, alt := range pref {
			if seen[alt] {
				return fmt.Errorf("int %d appears more than once in preference list %d", alt, i)
			}
			if !altMap[alt] {
				return fmt.Errorf("int %d is not in the int list", alt)
			}
			seen[alt] = true
		}
	}

	return nil
}

/*
Process of the vote

We distinguish between Social Welfare Functions (SWF),
which return a count based on a profile, and Social Choice Functions (social choice function, SCF)
which return only the preferred ints.
*/
// return the count for the given profile
func SWF(p Profile) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}

	count = make(Count)
	for _, pref := range p {
		for _, alt := range pref {
			count[alt]++
		}
	}
	return count, nil
}

// return the best ints for the given profile
func SCF(p Profile) (bestAlts []int, err error) {
	count, err := SWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}

// The simple Majority method
func MajoritySWF(p Profile) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}

	count = make(Count)
	for _, pref := range p {
		count[pref[0]]++
	}
	return count, nil
}

func MajoritySCF(p Profile) (bestAlts []int, err error) {
	count, err := MajoritySWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}

// The Borda method
func BordaSWF(p Profile) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}

	count = make(Count)
	for _, pref := range p {
		for i, alt := range pref {
			count[alt] += len(pref) - i - 1
		}
	}
	return count, nil
}

// 这个阈值是每个选民可以选择的最大的候选人数
func BordaSCF(p Profile) (bestAlts []int, err error) {
	count, err := BordaSWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}

// Approval voting method
func ApprovalSWF(p Profile, thresholds []int) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}
	//the thresholds is the maximum number of ints that each voter can choose
	//if the voter chooses more than the threshold, the ints beyond the threshold will be ignored
	//if the voter chooses less than the threshold, the ints he chooses will be counted

	//check if the length of the thresholds is equal to the length of the profile
	if len(thresholds) != len(p) {
		return nil, fmt.Errorf("the length of the thresholds is not equal to the length of the profile")
	}

	count = make(Count)
	for i, pref := range p {
		for j := 0; j < thresholds[i]; j++ {
			count[pref[j]]++
		}
	}
	return count, nil
}

func ApprovalSCF(p Profile, thresholds []int) (bestAlts []int, err error) {
	count, err := ApprovalSWF(p, thresholds)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}

// The tie-breaking method, which returns the best int among the given ints
// we use the random method to break the tie
func TieBreak(alts []int) (int, error) {
	if len(alts) == 0 {
		return 0, errors.New("empty ints list")
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Randomly select an int from the list
	selected := alts[rand.Intn(len(alts))]

	return selected, nil
}

//这里要我们创建多种TieBreak函数并实现工厂模式，我省略掉了

// SWFFactory is a function that returns a SWF
func SWFFactory(SWF func(p Profile) (Count, error), TieBreak func([]int) (int, error)) func(Profile) ([]int, error) {
	return func(p Profile) ([]int, error) {
		count, err := SWF(p)
		if err != nil {
			return nil, err
		}

		bestAlts := maxCount(count)
		if len(bestAlts) == 1 {
			return bestAlts, nil
		}

		if len(bestAlts) > 1 {
			bestAlt, err := TieBreak(bestAlts)
			if err != nil {
				return nil, err
			}
			return []int{bestAlt}, nil
		}

		//if there is no best int, return error
		return nil, fmt.Errorf("there is no best int")
	}
}

// SCFFactory is a function that returns a SCF

func SCFFactory(scf func(p Profile) ([]int, error), TieBreak func([]int) (int, error)) func(Profile) (int, error) {
	return func(p Profile) (int, error) {
		bestAlts, err := scf(p)
		if err != nil {
			return 0, err
		}
		if len(bestAlts) == 1 {
			return bestAlts[0], nil
		}
		if len(bestAlts) > 1 {
			bestAlt, err := TieBreak(bestAlts)
			if err != nil {
				return 0, err
			}
			return bestAlt, nil
		}
		return 0, fmt.Errorf("there is no best int")
	}
}

// To find the Condorcet winner, we need to compare each pair of ints.
// the return value is void or the Condorcet winner
func CondorcetWinner(p Profile) (bestAlt []int, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}

	numAlts := len(p[0])
	//the number of wins of each int
	wins := make([]int, numAlts)
	for i := 0; i < numAlts; i++ {
		for j := 0; j < numAlts; j++ {
			if i != j {
				winsForI := 0
				for _, voterPref := range p {
					if isPref(int(i+1), int(j+1), voterPref) {
						winsForI++
					}
				}
				if winsForI > len(p)/2 {
					wins[i]++
				}
			}
		}
	}
	fmt.Println(wins)
	//find the Condorcet winner
	for i, win := range wins {
		if win == numAlts-1 {
			bestAlt = append(bestAlt, int(i+1))
		}
	}

	return bestAlt, nil
}

// The Copeland method
// win +1, lose -1, tie 0
func CopelandSWF(p Profile) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}

	numAlts := len(p[0])
	count = make(Count)
	for i := 0; i < numAlts; i++ {
		for j := 0; j < numAlts; j++ {
			if i != j {
				winsForI := 0
				for _, voterPref := range p {
					if isPref(int(i), int(j), voterPref) {
						winsForI++
					}
				}
				if winsForI > len(p)/2 {
					count[int(i)]++
				} else if winsForI < len(p)/2 {
					count[int(i)]--
				}
			}
		}
	}

	return count, nil
}

func CopelandSCF(p Profile) (bestAlts []int, err error) {
	count, err := CopelandSWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}
