## Alien Invasion

Alien Invasion is a simulation program that creates configurable number of aliens and randomly places them across available number cities. At each step each alien will be randomly assigned to a new city until the total number of steps are exhausted (i.e 10000). when two aliens meet in the same city they destroy the city and disappear.

### How to build

``go build -o alien-invasion <path_to_alien_invasion>``

#### Example
 ``go build -o alien-invasion ./alien-invasion``

### How to Run

``./alien-invasion``

### How to see configuration

``./alien-invasion -h``

#### Example

```
Supported Fields:
FIELD          FLAG                   ENV                   DEFAULT
-----          -----                  -----                 -------
NumAliens      -num-aliens            NUM_ALIENS            5
CityMapFile    -city-map-file-path    CITY_MAP_FILE_PATH    input.txt

```
