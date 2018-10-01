package main

import (
    "math"
    "container/heap"
)

type Coord [2]float64
type Ring []Coord
type Polygon []Ring

type Cell struct {
    x float64
    y float64
    h float64
    d float64
    max float64
}

func NewCell(x float64, y float64, h float64, polygon Polygon) *Cell {
    d := pointToPolygonDistance(x, y, polygon)
    cell := Cell{x, y, h, d, d + h * math.Sqrt2}
    return &cell
}

func NewCellItem(cell *Cell) *Item {
    return &Item{cell, cell.d, 0}
}

func polylabel(polygon Polygon, precision float64) (float64, float64){
    minX, minY, maxX, maxY := boundingBox(polygon)
    
    width := maxX - minX
    height := maxY - minY
    cellSize := math.Min(width, height)
    h := cellSize / 2
    
    if cellSize == 0 {
        return minX, minY
    }
    
    cellQueue := make(PriorityQueue, 0)
    
    // cover polygon with initial cells
    for x:= minX; x < maxX; x += cellSize {
        for y := minY; y < maxY; y += cellSize {
            heap.Push(&cellQueue, NewCellItem(NewCell(x + h, y + h, h, polygon)))
        }
    }
    
    // take centroid as the first best guess
    bestCell := getCentroidCell(polygon)
    
    // special case for rectangular polygons
    bboxCell := NewCell(minX + width / 2, minY + height / 2, 0, polygon)
    if bboxCell.d > bestCell.d {
        bestCell = bboxCell
    }
    
    for cellQueue.Len() > 0 {
        // pick the most promising cell from the queue
        cellItem := heap.Pop(&cellQueue).(*Item)
        cell := cellItem.value
        
        // update the best cell if we found a better one
        if cell.d > bestCell.d {
            bestCell = cell
        }
        
        // do not drill down further if there's no chance of a better solution
        if (cell.max - bestCell.d) <= precision {
            continue
        }
        
        // split the cell into four cells
        h = cell.h / 2
        heap.Push(&cellQueue, NewCellItem(NewCell(cell.x - h, cell.y - h, h, polygon)))
        heap.Push(&cellQueue, NewCellItem(NewCell(cell.x + h, cell.y - h, h, polygon)))
        heap.Push(&cellQueue, NewCellItem(NewCell(cell.x - h, cell.y + h, h, polygon)))
        heap.Push(&cellQueue, NewCellItem(NewCell(cell.x + h, cell.y + h, h, polygon)))
    }
    
    return bestCell.x, bestCell.y
}

func boundingBox(polygon Polygon) (minX float64, minY float64, maxX float64, maxY float64){
    coords := polygon[0]
    minX, minY = coords[0][0], coords[0][1]
    maxX, maxY = coords[0][0], coords[0][1]
    for _, coord := range coords {
        x, y := coord[0], coord[1]
        if x < minX {
            minX = x
        }
        if x > maxX {
            maxX = x
        }
        if y < minY {
            minY = y
        }
        if y > maxY {
            maxY = y
        }
    }
    return
}

// signed distance from point to polygon outline (negative if point is outside)
func pointToPolygonDistance(x float64, y float64, polygon Polygon) float64 {
    inside := false
    minDistSq := math.Inf(1)
    
    for _, ring := range polygon {
        for n := 0; n < (len(ring) - 1); n++ {
            a := ring[n]
            b := ring[n + 1]
            if (((a[1] > y) != (b[1] > y)) && (x < ((b[0] - a[0]) * (y - a[1]) / (b[1] - a[1]) + a[0]))) {
                inside = !inside
            }
            minDistSq = math.Min(minDistSq, segmentDistanceSquared(x, y, a, b))
        }
    }
    
    factor := 1.0
    if !inside {
        factor = -1.0
    }
    return factor * math.Sqrt(minDistSq)
}

// get polygon centroid
func getCentroidCell(polygon Polygon) *Cell {
    area := 0.0
    x := 0.0
    y := 0.0
    ring := polygon[0]
    for n := 0; n < (len(ring) - 1); n++ {
        a := ring[n]
        b := ring[n + 1]
        f := a[0] * b[1] - b[0] * a[1]
        x += (a[0] + b[0]) * f
        y += (a[1] + b[1]) * f
        area += f * 3
    }
    if area == 0 {
        return NewCell(ring[0][0], ring[0][1], 0, polygon)
    }
    return NewCell(x / area, y / area, 0, polygon)
}

// get squared distance from a point to a segment
func segmentDistanceSquared(px float64, py float64, a [2]float64, b [2]float64) float64 {
    x := a[0]
    y := a[1]
    dx := b[0] - x
    dy := b[1] - y
    
    if dx != 0 || dy != 0 {
        t := ((px - x) * dx + (py - y) * dy) / (dx * dx + dy * dy)
        if t > 1 {
            x = b[0]
            y = b[1]
        } else if t > 0 {
            x += dx * t
            y += dy * t
        }
    }
    
    dx = px - x
    dy = py - y
    
    return dx * dx + dy * dy
}
