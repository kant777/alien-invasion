package main

import (
	"alien-invasion/conf"
	"bufio"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

type Direction string
type City string
type Alien string
type CityMap map[City]map[Direction]City

type Route struct {
	c City
	d Direction
}

const (
	North Direction = "north"
	South Direction = "south"
	East  Direction = "east"
	West  Direction = "west"
)

const MAX_STEPS = 10_000

func main() {
	config := conf.GetConfig()
	cityMap := parseCityMap(config.CityMapFile)
	aliensList := parseNames("names.txt")
	rand.Seed(time.Now().UnixNano())
	numAliens := config.NumAliens
	var aliens []Alien
	for i := 0; i < numAliens; i++ {
		aliens = append(aliens, aliensList[i])
	}
	initialAliens, numAlienSteps := initializeAliens(aliens, getAllCities(cityMap))
	log.Info().Msgf("initial city map: %v", cityMap)
	log.Info().Msgf("initialized aliens: %v", initialAliens)
	simulateAlienInvasion(initialAliens, numAlienSteps, cityMap)
}

/* This function simulates alien invasion by randomly picking directions
   from the city map for each alien from their initial positions at each step and
   when two or more aliens meet in the same city those aliens along with the city will disappear
   and the state of the simulation will be updated accordingly. The simulation also stops
   when all the aliens reach the MAX_STEPS
*/
func simulateAlienInvasion(initializedAliens map[Alien]City, numAlienSteps map[Alien]int, cityMap CityMap) {
	var updatedAliensInCity map[Alien]City
	var numAlienStepsSoFar map[Alien]int
	var updatedCityMap CityMap
	updatedAliensInCity, numAlienStepsSoFar, updatedCityMap = updateState(initializedAliens, numAlienSteps, cityMap)
	for checkIfAlienMovesExhausted(numAlienStepsSoFar) && len(updatedAliensInCity) > 0 && len(cityMap) > 0 {
		updatedAliensInCity, numAlienStepsSoFar = simulateNextStep(updatedAliensInCity, numAlienSteps, updatedCityMap)
		updatedAliensInCity, numAlienStepsSoFar, updatedCityMap = updateState(initializedAliens, numAlienSteps, cityMap)
		log.Info().Msgf("total steps by aliens: %v", numAlienSteps)
	}
}

/* This function randomly picks the next direction available from the city map for each Alien
   and returns the update state. It also increments total number of steps for each alien by 1
*/
func simulateNextStep(aliensInCity map[Alien]City, numAlienSteps map[Alien]int, cityMap CityMap) (map[Alien]City, map[Alien]int) {
	for alien, city := range aliensInCity {
		routeMap := cityMap[city]
		directions := getAllDirections(routeMap)
		if len(directions) > 0 {
			var directionIndex = 0
			if len(directions) > 0 {
				directionIndex = rand.Intn(len(directions))
			}
			nextDirection := directions[directionIndex]
			aliensInCity[alien] = routeMap[nextDirection]
		}
		numAlienSteps[alien] = numAlienSteps[alien] + 1
	}
	return aliensInCity, numAlienSteps
}

/* This function will update all the appropriate state after each step.
   @param aliensInCity tracks the current city for each Alien
   @param numAlienSteps tracks total number of steps taken by each alien
   @param cityMap tracks and update the graph of the city map
*/
func updateState(aliensInCity map[Alien]City, numAlienSteps map[Alien]int, cityMap CityMap) (map[Alien]City, map[Alien]int, CityMap) {
	aliensInSameCity := groupAliensInSameCity(aliensInCity)
	for city, aliensList := range aliensInSameCity {
		if len(aliensList) >= 2 {
			log.Info().Msgf("City %v has been destroyed by Aliens :%v", city, strings.Join(getAlienNames(aliensList), ","))
			cityMap = removeCity(city, cityMap)
			for _, alien := range aliensList {
				delete(aliensInCity, alien)
				delete(numAlienSteps, alien)
			}
		}
	}
	log.Info().Msgf("alien left in the city: %v", aliensInCity)
	log.Info().Msgf("updated city map: %v", cityMap)
	return aliensInCity, numAlienSteps, cityMap

}

/* Return the list of alien names */
func getAlienNames(aliens []Alien) []string {
	var names []string
	for _, alien := range aliens {
		names = append(names, string(alien))
	}
	return names
}

/* groups all the aliens the end up in the same city */
func groupAliensInSameCity(aliensInCity map[Alien]City) map[City][]Alien {
	groupAliensByCity := map[City][]Alien{}
	for alien, city := range aliensInCity {
		_, ok := groupAliensByCity[city]
		if ok {
			groupAliensByCity[city] = append(groupAliensByCity[city], alien)
		} else {
			groupAliensByCity[city] = []Alien{alien}
		}
	}
	return groupAliensByCity
}

/*Removes city from the city map. This includes removing both incoming and outgoing routes from the map*/
func removeCity(city City, cityMap CityMap) CityMap {
	for _, route := range cityMap {
		for d, c := range route {
			if city == c {
				delete(route, d)
			}
		}
	}
	delete(cityMap, city)
	return cityMap
}

/*checks to see if each alien had maxed out of steps */
func checkIfAlienMovesExhausted(alienSteps map[Alien]int) bool {
	for _, numSteps := range alienSteps {
		if numSteps < MAX_STEPS {
			return false
		}
	}
	return true
}

/*Randomly assign aliens to the available cities and updates the count of steps for each alien */
func initializeAliens(aliens []Alien, cities []City) (map[Alien]City, map[Alien]int) {
	alienCityMap := map[Alien]City{}
	numAlienSteps := map[Alien]int{}
	for i := 0; i < len(aliens); i++ {
		cityIndex := rand.Intn(len(cities))
		alien := aliens[i]
		alienCityMap[alien] = cities[cityIndex]
		numAlienSteps[alien] = 1
	}
	return alienCityMap, numAlienSteps
}

/*parses the city map from a file to CityMap structure */
func parseCityMap(path string) CityMap {
	file := open(path)
	defer close(file)

	cityMap := map[City]map[Direction]City{}
	scanner := bufio.NewScanner(file)
	// resize scanner's capacity for lines over 64K

	re := regexp.MustCompile(`\s+`)
	for scanner.Scan() {
		line := re.ReplaceAllString(scanner.Text(), " ")
		city, routes := parseLine(line)
		routeMap := map[Direction]City{}
		for _, r := range routes {
			routeMap[r.d] = r.c
		}
		cityMap[city] = routeMap
	}
	if err := scanner.Err(); err != nil {
		log.Fatal().Msgf("error reading a line from the file: %v", err)
	}
	return cityMap
}

func open(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal().Msgf("error opening the file: %v", err)
	}
	return file
}

func close(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Error().Msgf("error closing the file: %v", err)
	}
}

/*parse line expects each line in the following format to be parsed correctly
  Bar south=Foo west=Bee
*/
//TODO: Add a validate function
func parseLine(line string) (City, []Route) {
	cityRoutes := strings.Split(line, " ")
	var routes []Route
	for _, route := range cityRoutes[1:] {
		directionCityPair := strings.Split(route, "=")
		r := Route{
			d: Direction(strings.ToLower(directionCityPair[0])),
			c: City(strings.ToLower(directionCityPair[1])),
		}
		routes = append(routes, r)
	}
	return City(strings.ToLower(cityRoutes[0])), routes
}

func parseNames(path string) []Alien {
	file := open(path)
	defer close(file)

	var aliens []Alien
	scanner := bufio.NewScanner(file)
	// resize scanner's capacity for lines over 64K
	for scanner.Scan() {
		aliens = append(aliens, Alien(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal().Msgf("error reading name from the file: %v", err)
	}
	return aliens
}

func getAllCities(cityMap CityMap) []City {
	var cities []City
	for city := range cityMap {
		cities = append(cities, city)
	}
	return cities
}

func getAllDirections(routeMap map[Direction]City) []Direction {
	var directions []Direction
	for d, _ := range routeMap {
		directions = append(directions, d)
	}
	return directions
}
