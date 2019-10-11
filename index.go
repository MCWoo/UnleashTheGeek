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

const RADAR_DIST = 4
const MOVE_DIST = 4

/**********************************************************************************
 * Functions that the std library doesn't have
 *********************************************************************************/
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

/**********************************************************************************
 * Data structures
 *********************************************************************************/

type Map struct {
    width, height int
}

func (m Map) ArrayIndex(x, y int) int {
    return y * m.width + x
}

/**
 * A pair of ints for coordinates
 **/
type Coord struct {
    x, y int
}

func (c Coord) String() string {
    return fmt.Sprintf("(%d, %d)", c.x, c.y)
}

type Cmd int

const (
    CMD_WAIT  Cmd = 0
    CMD_MOVE  Cmd = 1
    CMD_DIG   Cmd = 2
    CMD_RADAR Cmd = 3
    CMD_TRAP  Cmd = 4
)

type Item int

const (
    ITEM_NONE  Item = -1
    ITEM_RADAR Item = 2
    ITEM_TRAP  Item = 3
    ITEM_ORE   Item = 4
)

type Object int

const (
    OBJ_ME       Object = 0
    OBJ_OPPONENT Object = 1
    OBJ_RADAR    Object = 2
    OBJ_TRAP     Object = 3
)

type Robot struct {
    id        int
    pos       Coord
    cmd       Cmd
    targetPos Coord
    item      Item
}

func (r *Robot) String() string {
    return fmt.Sprintf("Robot (%d) { pos: %s, cmd: %d, targetPos: %s, item: %d}", r.id, r.pos, r.cmd, r.targetPos, r.item)
}

func (r *Robot) Wait() {
    r.cmd = CMD_WAIT
}

func (r *Robot) Move(pos Coord) {
    r.cmd = CMD_MOVE
    r.targetPos.x = pos.x
    r.targetPos.y = pos.y
}

func (r *Robot) Dig(pos Coord) {
    r.cmd = CMD_DIG
    r.targetPos.x = pos.x
    r.targetPos.y = pos.y
}

func (r *Robot) RequestRadar() {
    r.cmd = CMD_RADAR
}

func (r *Robot) RequestTrap() {
    r.cmd = CMD_TRAP
}

func (r *Robot) GetCommand() string {
    if r.cmd == CMD_WAIT {
        return "WAIT"
    }
    if r.cmd == CMD_MOVE {
        return fmt.Sprintf("MOVE %d %d", r.targetPos.x, r.targetPos.y)
    }
    if r.cmd == CMD_DIG {
        return fmt.Sprintf("DIG %d %d", r.targetPos.x, r.targetPos.y)
    }
    if r.cmd == CMD_RADAR {
        return "REQUEST RADAR"
    }
    if r.cmd == CMD_TRAP {
        return "REQUEST TRAP"
    }
    fmt.Fprintf(os.Stderr, "Unknown command type for robot! %d, id: %d", r.cmd, r.id)
    return "WAIT"
}

/**********************************************************************************
 * Utility functions
 *********************************************************************************/

/**
 * The Manhattan distance between 2 coordinates
 **/
func dist(p1, p2 Coord) int {
    return abs(p1.x-p2.x) + abs(p1.y-p2.y)
}

/**
 * The Manhattan distance between 2 coordinates for digging (1 less)
 **/
func digDist(p1, p2 Coord) int {
    return max(abs(p1.x-p2.x)+abs(p1.y-p2.y)-1, 0)
}

/**
 * The distance in turns between 2 coordinates
 **/
func turnDist(p1, p2 Coord) int {
    return int(math.Ceil(float64(dist(p1, p2)) / MOVE_DIST))
}

/**
 * The distance in turns between 2 coordinates for digging
 **/
func digTurnDist(p1, p2 Coord) int {
    return int(math.Ceil(float64(digDist(p1, p2)) / MOVE_DIST))
}

/**********************************************************************************
 * Serious business here
 *********************************************************************************/
func calculateCellRadarValues(unknowns []int, width, height int) []int {
    radarValues := make([]int, width*height)
    for j := 0; j < height; j++ {
        for i := 0; i < width; i++ {
            cell := Coord{i, j}
            for n := max(j-RADAR_DIST, 0); n < min(j+RADAR_DIST, height); n++ {
                for m := max(i-4, RADAR_DIST); m < min(i+RADAR_DIST, width); m++ {
                    if dist(cell, Coord{m, n}) > RADAR_DIST {
                        continue
                    }
                    radarValues[cell.y*width+cell.x] += unknowns[n*width+m]
                }
            }
        }
    }
    return radarValues
}

func calculateBestRadarPosition(unknowns []int, width, height int, pos Coord) (best Coord) {
    radarValues := calculateCellRadarValues(unknowns, width, height)
    closest := width + height // furthest point
    largestValue := 0         // lowest value

    for j := 0; j < height; j++ {
        for i := 0; i < width; i++ {
            value := radarValues[j*width+i]
            if value > largestValue {
                largestValue = value
                best = Coord{i, j}
                closest = i
            } else if value == largestValue {
                newCoord := Coord{i, j}
                // Pick the closest to HQ
                if i < closest {
                    best = newCoord
                    closest = i
                }
            }
        }
    }
    return best
}

/**********************************************************************************
 * Main loop
 *********************************************************************************/
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Buffer(make([]byte, 1000000), 1000000)

    // height: size of the map
    var width, height int
    scanner.Scan()
    fmt.Sscan(scanner.Text(), &width, &height)
    ores := make([]int, width*height)
    unknowns := make([]int, width*height)
    robots := make([]Robot, 5)

    for {
        start := time.Now()
        // myScore: Amount of ore delivered
        var myScore, opponentScore int
        scanner.Scan()
        fmt.Sscan(scanner.Text(), &myScore, &opponentScore)

        for i := 0; i < height; i++ {
            scanner.Scan()
            inputs := strings.Split(scanner.Text(), " ")
            for j := 0; j < width; j++ {
                // ore: amount of ore or "?" if unknown
                // hole: 1 if cell has a hole
                ore, err := strconv.Atoi(inputs[2*j])
                if err != nil {
                    ores[i*width+j] = 0
                    unknowns[i*width+j] = 1
                } else {
                    ores[i*width+j] = ore
                    unknowns[i*width+j] = 0
                }

                hole, _ := strconv.ParseInt(inputs[2*j+1], 10, 32)
                _ = hole
            }
        }

        // entityCount: number of entities visible to you
        // radarCooldown: turns left until a new radar can be requested
        // trapCooldown: turns left until a new trap can be requested
        var entityCount, radarCooldown, trapCooldown int
        scanner.Scan()
        fmt.Sscan(scanner.Text(), &entityCount, &radarCooldown, &trapCooldown)
        myRobot_i := 0
        for i := 0; i < entityCount; i++ {
            // id: unique id of the entity
            // type: 0 for your robot, 1 for other robot, 2 for radar, 3 for trap
            // y: position of the entity
            // item: if this entity is a robot, the item it is carrying (-1 for NONE, 2 for RADAR, 3 for TRAP, 4 for ORE)
            var id, objType, x, y, item int
            scanner.Scan()
            fmt.Sscan(scanner.Text(), &id, &objType, &x, &y, &item)

            if Object(objType) == OBJ_ME {
                robot := &robots[myRobot_i]
                robot.id = id
                robot.pos.x = x
                robot.pos.y = y
                robot.item = Item(item)

                myRobot_i++
            } else if Object(objType) == OBJ_TRAP {
                ores[y*width + x] = 0
            }
        }

        chosenWidth := width
        for j := 0; j < height; j++ {
            for i := 0; i < width; i++ {
                if ores[j*width + i] != 0 && j < chosenWidth {
                    chosenWidth = i
                    for robots_i := 1; robots_i < len(robots); robots_i++ {
                        robots[robots_i].Dig(Coord{i, j})
                    }
                }
            }
        }

        for i := 0; i < len(robots); i++ {
            robot := &robots[i]
            if i == 0 {
                if robot.item == ITEM_NONE {
                    robot.RequestRadar()
                } else if robot.pos.x == 0 {
                    bestDig := calculateBestRadarPosition(unknowns, width, height, robot.pos)
                    robot.Dig(bestDig)
                }
            } else {
                if robot.item == ITEM_ORE {
                    robot.Move(Coord{0, robot.pos.y})
                }
            }
        }

        for i := 0; i < len(robots); i++ {
            fmt.Println(robots[i].GetCommand()) // WAIT|MOVE x y|DIG x y|REQUEST item
        }
        elapsed := time.Since(start)
        fmt.Fprintf(os.Stderr, "%v elapsed in turn", elapsed)
    }
}
