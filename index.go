package main

import (
    "fmt"
    "time"
)
import "os"
import "bufio"
import "strings"
import "strconv"
import "math"
//import "math/rand"

const RADAR_DIST = 4
const MOVE_DIST = 4

func abs(n int) int {
	if n < 0 {
	    return -n
    }
    return n
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

/**
 * A pair of ints for coordinates
 **/
type coord struct {
    x, y int
}

/**
 * The Manhattan distance between 2 coordinates
 **/
func dist(p1, p2 coord) int {
    return abs(p1.x - p2.x) + abs(p1.y - p2.y)
}

/**
 * The Manhattan distance between 2 coordinates for digging (1 less)
 **/
func digDist(p1, p2 coord) int {
    return max(abs(p1.x - p2.x) + abs(p1.y - p2.y) - 1, 0)
}

/**
 * The distance in turns between 2 coordinates
 **/
func turnDist(p1, p2 coord) int {
    return int(math.Ceil(float64(dist(p1, p2)) / MOVE_DIST))
}

/**
 * The distance in turns between 2 coordinates for digging
 **/
func digTurnDist(p1, p2 coord) int {
    return int(math.Ceil(float64(digDist(p1, p2)) / MOVE_DIST))
}

func calculateCellRadarValues(unknowns []int, width, height int) []int {
	radarValues := make([]int, width*height)
    for j := 0; j < height; j++ {
    	for i := 0; i < width; i++ {
    	    cell := coord{i, j}
    	    for n := max(j-RADAR_DIST, 0); n < min(j+RADAR_DIST, height); n++ {
                for m := max(i-4, RADAR_DIST); m < min(i+RADAR_DIST, width); m++ {
                    if dist(cell, coord{m, n}) > RADAR_DIST {
                        continue
                    }
                    radarValues[cell.y*width + cell.x] += unknowns[n*width + m]
                }
            }
        }
    }
    return radarValues
}

func calculateBestRadarPosition(unknowns []int, width, height int, pos coord) (best coord) {
    radarValues := calculateCellRadarValues(unknowns, width, height)
    closest := width + height // furthest point
    largestValue := 0 // lowest value

    for j := 0; j < height; j++ {
        for i := 0; i < width; i++ {
            value := radarValues[j*width + i]
            if value > largestValue {
                largestValue = value
                best = coord{i, j}
                closest = i //digTurnDist(pos, best)
            } else if value == largestValue {
                newCoord := coord{i, j}
                dist := i //digTurnDist(pos, newCoord)
                if dist < closest {
                    best = newCoord
                    closest = dist
                }
            }
        }
    }
    return best
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Buffer(make([]byte, 1000000), 1000000)

    // height: size of the map
    var width, height int
    scanner.Scan()
    fmt.Sscan(scanner.Text(),&width, &height)
    ores := make([]int, width*height)
    unknowns := make([]int, width*height)
    cmds := []string{"WAIT", "WAIT", "WAIT", "WAIT", "WAIT"}
    for {
    	start := time.Now()
        // myScore: Amount of ore delivered
        var myScore, opponentScore int
        scanner.Scan()
        fmt.Sscan(scanner.Text(),&myScore, &opponentScore)
        
        chosenWidth := width
        for i := 0; i < height; i++ {
            scanner.Scan()
            inputs := strings.Split(scanner.Text()," ")
            for j := 0; j < width; j++ {
                // ore: amount of ore or "?" if unknown
                // hole: 1 if cell has a hole
                ore, err := strconv.Atoi(inputs[2*j])
                if err != nil {
                    ores[i*width + j] = 0
                    unknowns[i*width + j] = 1
                } else {
                    ores[i*width + j] = ore
                    unknowns[i*width + j] = 0
                }
                
                if ore != 0 && j < chosenWidth {
                    chosenWidth = j
                    for cmd_i := 1; cmd_i < len(cmds); cmd_i++ {
                        cmds[cmd_i] = fmt.Sprintf("DIG %d %d", j, i)
                    }
                }
                hole,_ := strconv.ParseInt(inputs[2*j+1],10,32)
                _ = hole
            }
        }
        
        // for cmd_i := 1; cmd_i < len(cmds); cmd_i++ {
        //     cmds[cmd_i] = fmt.Sprintf("DIG %d %d", rand.Intn(width), rand.Intn(height))
        // }
        
        // entityCount: number of entities visible to you
        // radarCooldown: turns left until a new radar can be requested
        // trapCooldown: turns left until a new trap can be requested
        var entityCount, radarCooldown, trapCooldown int
        scanner.Scan()
        fmt.Sscan(scanner.Text(),&entityCount, &radarCooldown, &trapCooldown)
        myRobot_i := 0
        for i := 0; i < entityCount; i++ {
            // id: unique id of the entity
            // type: 0 for your robot, 1 for other robot, 2 for radar, 3 for trap
            // y: position of the entity
            // item: if this entity is a robot, the item it is carrying (-1 for NONE, 2 for RADAR, 3 for TRAP, 4 for ORE)
            var id, objType, x, y, item int
            scanner.Scan()
            fmt.Sscan(scanner.Text(),&id, &objType, &x, &y, &item)
            
            if objType == 0 {
                
                if myRobot_i == 0 {
                    if item == -1 {
                        cmds[0] = "REQUEST RADAR"
                    } else if x == 0 {
                        bestDig := calculateBestRadarPosition(unknowns, width, height, coord{x, y})
                        cmds[0] = fmt.Sprintf("DIG %d %d", bestDig.x, bestDig.y)
                    }
                } else {
                    if item == 4 {
                        cmds[myRobot_i] = fmt.Sprintf("MOVE %d %d", 0, y)
                    }
                }
                myRobot_i++
            }
        }
        for i := 0; i < 5; i++ {
            
            // fmt.Fprintln(os.Stderr, "Debug messages...")
            fmt.Println(cmds[i]) // WAIT|MOVE x y|DIG x y|REQUEST item
        }
        elapsed := time.Since(start)
        fmt.Fprintf(os.Stderr, "%v elapsed in turn", elapsed)
    }
}