package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type EdgeType int

const (
	Normal EdgeType = iota
	Closed
	Boosting
	Barrier
	Magnet
)

type duo struct {
	EndPoint int
	Weight   int
}

type duoPath struct {
	PrevPoint int
	EdgeType  int
}

type trio struct {
	EndPoint int
	Weight   int
	EdgeType
}

type fourths struct {
	StartPoint int
	EndPoint   int
	Weight     int
	EdgeType
}

func Contains(a []int, x int) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func MakeSimpleGraph(graph [][]trio) [][]duo {
	n := len(graph)
	var simpleGraph = make([][]duo, n)
	for i, v := range graph {
		for _, e := range v {
			simpleGraph[i] = append(simpleGraph[i], duo{e.EndPoint, e.Weight})
		}
	}
	return simpleGraph
}

func DeleteExcessEdges(graph [][]trio) [][]trio {
	newGraph := make([][]trio, len(graph))
	for j, v := range graph {
		m := make(map[trio]*trio, len(v))
		for i, edge := range v {
			edge.Weight = 0
			if w := m[edge]; w == nil || w.Weight > v[i].Weight {
				m[edge] = &v[i]
			}
		}
		for _, e := range m {
			newGraph[j] = append(newGraph[j], *e)
		}
	}
	return newGraph
}

func MakeAuxiliaryGraphForMix(graph [][]trio) [][]duo {
	n := len(graph)
	var auxGraph = make([][]duo, 2*n)
	for i, v := range graph {
		for _, e := range v {
			switch e.EdgeType {
			case Normal:
				auxGraph[i] = append(auxGraph[i], duo{e.EndPoint, e.Weight})
				auxGraph[i+n] = append(auxGraph[i+n], duo{e.EndPoint, e.Weight})
			case Closed:
				auxGraph[i] = append(auxGraph[i], duo{e.EndPoint + n, e.Weight})
			default:
				panic(fmt.Sprintf("%v", i))
			}
		}
	}
	return auxGraph
}

func MakeAuxiliaryGraphForBarrier(graph [][]trio, barlevel int) [][]duo {
	n := len(graph)
	var auxGraph = make([][]duo, (barlevel+1)*n)
	for j := 0; j <= barlevel; j++ {
		for i, v := range graph {
			for _, e := range v {
				switch e.EdgeType {
				case Normal:
					auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
				case Boosting:
					if j < barlevel {
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*(j+1), e.Weight})
					} else {
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					}

				case Barrier:
					if j >= barlevel {
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint, e.Weight})
					}
				default:
					panic(fmt.Sprintf("%v", i))
				}
			}
		}
	}
	return auxGraph
}

func MakeAuxiliaryGraphForMagnetBarrier(graph [][]trio, maglevel int) [][]duo {
	n := len(graph)
	var auxGraph = make([][]duo, (maglevel+1)*n)
	for j := 0; j <= maglevel; j++ {
		for i, v := range graph {
			if j < maglevel {
				for _, e := range v {
					switch e.EdgeType {
					case Normal:
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					case Boosting:
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*(j+1), e.Weight})
					case Magnet:
					default:
						panic(fmt.Sprintf("%v", i))
					}
				}
			} else {
				var containsMagnetEdges = false
				for _, e := range v {
					if e.EdgeType == 4 {
						containsMagnetEdges = true
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					}
				}
				if !containsMagnetEdges {
					for _, e := range v {
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					}
				}
			}
		}
	}
	return auxGraph
}

func MakeAuxiliaryGraphForMagnet(graph [][]trio, maglevel int) [][]duo {
	n := len(graph)
	var auxGraph = make([][]duo, (maglevel+1)*n)
	for j := 0; j <= maglevel; j++ {
		for i, v := range graph {
			if j < maglevel {
				for _, e := range v {
					switch e.EdgeType {
					case Normal:
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					case Boosting:
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*(j+1), e.Weight})
					case Magnet:
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					default:
						panic(fmt.Sprintf("%v", i))
					}
				}
			} else {
				var containsMagnetEdges = false
				for _, e := range v {
					if e.EdgeType == 4 {
						containsMagnetEdges = true
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*(j-1), e.Weight})
					}
				}
				if !containsMagnetEdges {
					for _, e := range v {
						auxGraph[i+n*j] = append(auxGraph[i+n*j], duo{e.EndPoint + n*j, e.Weight})
					}
				}
			}
		}
	}
	return auxGraph
}

func FindWeightInAuxGraph(auxGraph [][]duo, index int, pathInd int) int {
	n := len(auxGraph[index])
	for i := 0; i < n; i++ {
		if auxGraph[index][i].EndPoint == pathInd {
			return auxGraph[index][i].Weight
		}
	}
	return 0
}

func FindEdgeTypeInGraph(graph [][]trio, index int, pathInd int) EdgeType {
	n := len(graph[index])
	for i := 0; i < n; i++ {
		if graph[index][i].EndPoint == pathInd {
			return graph[index][i].EdgeType
		}
	}
	return 0
}

func MakeSourcePathForMix(path []int, auxGraph [][]duo, lenGraph int) []fourths {
	truePath := make([]fourths, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		if path[i] >= lenGraph {
			truePath = append(truePath,
				fourths{path[i] - lenGraph, path[i+1], FindWeightInAuxGraph(auxGraph, path[i], path[i+1]), 0})
		} else {
			if path[i+1] >= lenGraph {
				truePath = append(truePath,
					fourths{path[i], path[i+1] - lenGraph, FindWeightInAuxGraph(auxGraph, path[i], path[i+1]), 1})
			} else {
				truePath = append(truePath,
					fourths{path[i], path[i+1], FindWeightInAuxGraph(auxGraph, path[i], path[i+1]), 0})
			}
		}
	}
	return truePath
}

func MakeSourcePathForBarrier(path []int, graph [][]trio, auxGraph [][]duo, lenGraph int) []fourths {
	truePath := make([]fourths, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		truePath = append(truePath,
			fourths{path[i] % lenGraph, path[i+1] % lenGraph, FindWeightInAuxGraph(auxGraph, path[i], path[i+1]),
				FindEdgeTypeInGraph(graph, path[i]%lenGraph, path[i+1]%lenGraph)})
	}
	return truePath
}

func DeijkstraAlgorithm(graph [][]duo, startPoint int, finishPoint int) ([]int, int) {
	n := len(graph)
	dists, labels, prevPoints := make([]int, n), make([]bool, n), make([]int, n)
	path := make([]int, 0, n)
	for i := 1; i < n; i++ {
		dists[i] = int(^uint(0) >> 1)
	}
	for i := 0; i < n; i++ {
		v := -1

		for j := 0; j < n; j++ {
			if !labels[j] && (v == -1 || dists[j] < dists[v]) {
				v = j
			}
		}
		if dists[v] == int(^uint(0)>>1) {
			break
		}
		labels[v] = true

		for j := 0; j < len(graph[v]); j++ {
			to, length := graph[v][j].EndPoint, graph[v][j].Weight
			if dists[v]+length < dists[to] {
				dists[to] = dists[v] + length
				prevPoints[to] = v
			}
		}
	}
	for v := finishPoint; v != startPoint; v = prevPoints[v] {
		path = append([]int{v}, path...)
	}
	path = append([]int{startPoint}, path...)
	return path, dists[finishPoint]
}

func DeijkstraAlgorithmForAuxGraph(graph [][]duo, startPoint int, finishPoint int, limitlevel int, lenSourceGraph int) ([]int, int) {
	n := len(graph)
	dists, labels, prevPoints := make([]int, n), make([]bool, n), make([]int, n)
	path := make([]int, 0, n)
	for i := 1; i < n; i++ {
		dists[i] = int(^uint(0) >> 1)
	}
	for i := 0; i < n; i++ {
		v := -1

		for j := 0; j < n; j++ {
			if !labels[j] && (v == -1 || dists[j] < dists[v]) {
				v = j
			}
		}
		if dists[v] == int(^uint(0)>>1) {
			break
		}
		labels[v] = true

		for j := 0; j < len(graph[v]); j++ {
			to, length := graph[v][j].EndPoint, graph[v][j].Weight
			if dists[v]+length < dists[to] {
				dists[to] = dists[v] + length
				prevPoints[to] = v
			}
		}
	}

	minFinishPoint := finishPoint
	minDist := dists[minFinishPoint]
	for i := 1; i <= limitlevel; i++ {
		if dists[finishPoint+i*lenSourceGraph] < minDist {
			minFinishPoint = finishPoint + i*lenSourceGraph
			minDist = dists[finishPoint+i*lenSourceGraph]
		}
	}

	for v := minFinishPoint; v != startPoint; v = prevPoints[v] {
		path = append([]int{v}, path...)
	}
	path = append([]int{startPoint}, path...)
	return path, dists[minFinishPoint]
}

func DeijkstraVectorAlgorithmForBarrier(graph [][]trio, startPoint int, finishPoint int, barlevel int) ([]int, int) {
	defer func(t time.Time) {
		fmt.Println("Время работы векторного алгоритма Дейкстры", time.Since(t))
	}(time.Now())
	n := len(graph)
	dists, labels, prevPoints := make([][]int, n), make([][]bool, n), make([][]duoPath, n)
	currLevel := 0
	path := make([]int, 0, n)
	for i := range dists {
		dists[i] = make([]int, barlevel+1)
		labels[i] = make([]bool, barlevel+1)
		prevPoints[i] = make([]duoPath, barlevel+1)
	}

	for j := 1; j <= barlevel; j++ {
		dists[0][j] = int(^uint(0) >> 1)
	}

	for i := 1; i < n; i++ {
		for j := 0; j <= barlevel; j++ {
			dists[i][j] = int(^uint(0) >> 1)
		}
	}

	for z := 0; z <= (barlevel+1)*n; z++ {
		v := -1
		firstPoint := true
		for q := 0; q <= barlevel; q++ {
			for j := 0; j < n; j++ {
				if firstPoint {
					if !labels[j][q] && (v == -1 || dists[j][q] < dists[v][q]) {
						v = j
						currLevel = q
						firstPoint = false
					}
				} else {
					if !labels[j][q] && (v == -1 || dists[j][q] < dists[v][currLevel]) {
						v = j
						currLevel = q
					}
				}
			}
		}

		if dists[v][currLevel] == int(^uint(0)>>1) {
			break
		}

		labels[v][currLevel] = true

		for _, g := range graph[v] {
			to, length, edgeType := g.EndPoint, g.Weight, g.EdgeType
			switch edgeType {
			case Normal:
				if dists[v][currLevel]+length < dists[to][currLevel] {
					dists[to][currLevel] = dists[v][currLevel] + length
					prevPoints[to][currLevel].PrevPoint = v
					prevPoints[to][currLevel].EdgeType = 0
				}
			case Boosting:
				if currLevel < barlevel {
					if dists[v][currLevel]+length < dists[to][currLevel+1] {
						dists[to][currLevel+1] = dists[v][currLevel] + length
						prevPoints[to][currLevel+1].PrevPoint = v
						prevPoints[to][currLevel+1].EdgeType = 2
					}
				} else {
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 0
					}
				}
			case Barrier:
				if currLevel >= barlevel {
					if dists[v][currLevel]+length < dists[to][0] {
						dists[to][0] = dists[v][currLevel] + length
						prevPoints[to][0].PrevPoint = v
						prevPoints[to][0].EdgeType = 3
					}
				}
			}
		}
	}

	minDistLevel := 0
	minDist := dists[finishPoint][0]
	for i := 1; i <= barlevel; i++ {
		if dists[finishPoint][i] < minDist {
			minDist = dists[finishPoint][i]
			minDistLevel = i
		}
	}
	k := minDistLevel
	v := finishPoint
	for v != startPoint || k != 0 {
		path = append([]int{v}, path...)
		switch prevPoints[v][k].EdgeType {
		case 0:
			v = prevPoints[v][k].PrevPoint
		case 2:
			v = prevPoints[v][k].PrevPoint
			k--

		case 3:
			v = prevPoints[v][k].PrevPoint
			k = barlevel
		}
	}
	path = append([]int{startPoint}, path...)

	return path, dists[finishPoint][minDistLevel]
}

func DeijkstraVectorAlgorithmForMagnet(graph [][]trio, startPoint int, finishPoint int, maglevel int) ([]int, int) {
	defer func(t time.Time) {
		fmt.Println("Время работы векторного алгоритма:", time.Since(t))
	}(time.Now())
	n := len(graph)
	dists, labels, prevPoints := make([][]int, n), make([][]bool, n), make([][]duoPath, n)
	currLevel := 0
	path := make([]int, 0, n)
	for i := range dists {
		dists[i] = make([]int, maglevel+1)
		labels[i] = make([]bool, maglevel+1)
		prevPoints[i] = make([]duoPath, maglevel+1)
	}
	for j := 1; j <= maglevel; j++ {
		dists[0][j] = int(^uint(0) >> 1)
	}
	for i := 1; i < n; i++ {
		for j := 0; j <= maglevel; j++ {
			dists[i][j] = int(^uint(0) >> 1)
		}
	}

	for z := 0; z <= (maglevel+1)*n; z++ {
		v := -1
		firstPoint := true
		for q := 0; q <= maglevel; q++ {
			for j := 0; j < n; j++ {
				if firstPoint {
					if !labels[j][q] && (v == -1 || dists[j][q] < dists[v][q]) {
						v = j
						currLevel = q
						firstPoint = false
					}
				} else {
					if !labels[j][q] && (v == -1 || dists[j][q] < dists[v][currLevel]) {
						v = j
						currLevel = q
					}
				}
			}
		}

		if dists[v][currLevel] == int(^uint(0)>>1) {
			break
		}

		labels[v][currLevel] = true

		if currLevel == maglevel {
			var containsMagnetEdges = false
			for _, g := range graph[v] {
				to, length, edgeType := g.EndPoint, g.Weight, g.EdgeType
				if edgeType == 4 {
					containsMagnetEdges = true
					if dists[v][currLevel]+length < dists[to][currLevel-1] {
						dists[to][currLevel-1] = dists[v][currLevel] + length
						prevPoints[to][currLevel-1].PrevPoint = v
						prevPoints[to][currLevel-1].EdgeType = 4
					}
				}
			}
			if !containsMagnetEdges {
				for _, g := range graph[v] {
					to, length := g.EndPoint, g.Weight
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 0
					}
				}
			}
		} else {
			for _, g := range graph[v] {
				to, length, edgeType := g.EndPoint, g.Weight, g.EdgeType
				switch edgeType {
				case Normal:
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 0
					}
				case Boosting:
					if dists[v][currLevel]+length < dists[to][currLevel+1] {
						dists[to][currLevel+1] = dists[v][currLevel] + length
						prevPoints[to][currLevel+1].PrevPoint = v
						prevPoints[to][currLevel+1].EdgeType = 2
					}
				case Magnet:
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 0
					}
				}
			}
		}
	}
	minDistLevel := 0
	minDist := dists[finishPoint][0]
	for i := 1; i <= maglevel; i++ {
		if dists[finishPoint][i] < minDist {
			minDist = dists[finishPoint][i]
			minDistLevel = i
		}
	}
	k := minDistLevel
	v := finishPoint
	for v != startPoint || k != 0 {
		path = append([]int{v}, path...)
		switch prevPoints[v][k].EdgeType {
		case 0:
			v = prevPoints[v][k].PrevPoint
		case 2:
			v = prevPoints[v][k].PrevPoint
			k--
		case 4:
			v = prevPoints[v][k].PrevPoint
			k++
		}
	}
	path = append([]int{startPoint}, path...)
	return path, dists[finishPoint][minDistLevel]
}

func DeijkstraVectorAlgorithmForMagnetBarrier(graph [][]trio, startPoint int, finishPoint int, maglevel int) ([]int, int) {
	defer func(t time.Time) {
		fmt.Println("Время работы векторного алгоритма", time.Since(t))
	}(time.Now())
	n := len(graph)
	dists, labels, prevPoints := make([][]int, n), make([][]bool, n), make([][]duoPath, n)
	currLevel := 0
	path := make([]int, 0, n)
	for i := range dists {
		dists[i] = make([]int, maglevel+1)
		labels[i] = make([]bool, maglevel+1)
		prevPoints[i] = make([]duoPath, maglevel+1)
	}

	for j := 1; j <= maglevel; j++ {
		dists[0][j] = int(^uint(0) >> 1)
	}

	for i := 1; i < n; i++ {
		for j := 0; j <= maglevel; j++ {
			dists[i][j] = int(^uint(0) >> 1)
		}
	}

	for z := 0; z <= (maglevel+1)*n; z++ {
		v := -1
		firstPoint := true
		for q := 0; q <= maglevel; q++ {
			for j := 0; j < n; j++ {
				if firstPoint {
					if !labels[j][q] && (v == -1 || dists[j][q] < dists[v][q]) {
						v = j
						currLevel = q
						firstPoint = false
					}
				} else {
					if !labels[j][q] && (v == -1 || dists[j][q] < dists[v][currLevel]) {
						v = j
						currLevel = q
					}
				}
			}
		}

		if dists[v][currLevel] == int(^uint(0)>>1) {
			break
		}

		labels[v][currLevel] = true

		if currLevel == maglevel {
			var containsMagnetEdges = false
			for _, g := range graph[v] {
				to, length, edgeType := g.EndPoint, g.Weight, g.EdgeType
				if edgeType == 4 {
					containsMagnetEdges = true
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 4
					}
				}
			}
			if !containsMagnetEdges {
				for _, g := range graph[v] {
					to, length := g.EndPoint, g.Weight
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 0
					}
				}
			}
		} else {
			for _, g := range graph[v] {
				to, length, edgeType := g.EndPoint, g.Weight, g.EdgeType
				switch edgeType {
				case Normal:
					if dists[v][currLevel]+length < dists[to][currLevel] {
						dists[to][currLevel] = dists[v][currLevel] + length
						prevPoints[to][currLevel].PrevPoint = v
						prevPoints[to][currLevel].EdgeType = 0
					}
				case Boosting:
					if dists[v][currLevel]+length < dists[to][currLevel+1] {
						dists[to][currLevel+1] = dists[v][currLevel] + length
						prevPoints[to][currLevel+1].PrevPoint = v
						prevPoints[to][currLevel+1].EdgeType = 2
					}
				}
			}
		}
	}

	minDistLevel := 0
	minDist := dists[finishPoint][0]
	for i := 1; i <= maglevel; i++ {
		if dists[finishPoint][i] < minDist {
			minDist = dists[finishPoint][i]
			minDistLevel = i
		}
	}
	k := minDistLevel
	v := finishPoint
	for v != startPoint || k != 0 {
		path = append([]int{v}, path...)
		switch prevPoints[v][k].EdgeType {
		case 0:
			v = prevPoints[v][k].PrevPoint
		case 2:
			v = prevPoints[v][k].PrevPoint
			k--
		case 4:
			v = prevPoints[v][k].PrevPoint
		}
	}
	path = append([]int{startPoint}, path...)
	return path, dists[finishPoint][minDistLevel]
}

func ReadGraphForMix(filename string) [][]trio {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n, m, k := 0, 0, 0
	fmt.Fscanln(f, &n, &m, &k)
	fmt.Println("число вершин в графе -", n)
	fmt.Println("число дуг графа -", m)
	fmt.Println("число запрещенных дуг -", k)
	graph := make([][]trio, n)
	fmt.Println("Список дуг:")
	for i := 0; i < m; i++ {
		edge := trio{}
		var vertex int
		fmt.Fscanln(f, &vertex, &edge.EndPoint, &edge.Weight, &edge.EdgeType)
		fmt.Println(edge)
		graph[vertex] = append(graph[vertex], edge)
	}
	return graph
}

func ReadGraphForBarrier(filename string) ([][]trio, int) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n, m, barlevel := 0, 0, 0
	fmt.Fscanln(f, &n, &m, &barlevel)
	fmt.Println("число вершин в графе -", n)
	fmt.Println("число дуг графа -", m)
	fmt.Println("уровень барьера -", barlevel)
	graph := make([][]trio, n)
	fmt.Println("Список дуг:")
	for i := 0; i < m; i++ {
		edge := trio{}
		var vertex int
		fmt.Fscanln(f, &vertex, &edge.EndPoint, &edge.Weight, &edge.EdgeType)
		fmt.Println(edge)
		graph[vertex] = append(graph[vertex], edge)
	}
	return graph, barlevel
}

func ReadGraphForMagnet(filename string) ([][]trio, int) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n, m, maglevel := 0, 0, 0
	fmt.Fscanln(f, &n, &m, &maglevel)
	fmt.Println("число вершин в графе -", n)
	fmt.Println("число дуг графа -", m)
	fmt.Println("уровень магнитности -", maglevel)
	graph := make([][]trio, n)
	fmt.Println("Список дуг:")
	for i := 0; i < m; i++ {
		edge := trio{}
		var vertex int
		fmt.Fscanln(f, &vertex, &edge.EndPoint, &edge.Weight, &edge.EdgeType)
		fmt.Println(edge)
		graph[vertex] = append(graph[vertex], edge)
	}
	return graph, maglevel
}

func MixProgramm() {
	var graph = ReadGraphForMix("Graph1.txt")
	fmt.Println("Список смежности графа:")
	for i := 0; i < len(graph); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graph[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graph[i]); j++ {
			fmt.Println(graph[i][j])
		}
	}
	fmt.Println()

	//граф после удаления лишних дуг
	graph = DeleteExcessEdges(graph)
	fmt.Println("Список смежности графа после удаления лишних дуг:")
	for i := 0; i < len(graph); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graph[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graph[i]); j++ {
			fmt.Println(graph[i][j])
		}
	}
	//fmt.Println(graph)
	fmt.Println()

	//Вспомогательный граф
	var auxGraph = MakeAuxiliaryGraphForMix(graph)
	fmt.Println("Список смежности вспомогательного графа:")
	for i := 0; i < len(auxGraph); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(auxGraph[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(auxGraph[i]); j++ {
			fmt.Println(auxGraph[i][j])
		}
	}
	//fmt.Println(auxGraph)
	fmt.Println()

	//алгоритм Дейкстры
	var simpleGraph = MakeSimpleGraph(graph)
	var startPointSimple, finishPointSourceSimple = 0, 5
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа без ограничений для пути,")
	fmt.Println("который начинается в вершине", startPointSimple, "и заканчивается в вершине", finishPointSourceSimple, ":")
	fmt.Println("Результат для графа без ограничений:")
	var pathSourceSimple, distSourceSimple = DeijkstraAlgorithm(simpleGraph, startPointSimple, finishPointSourceSimple)
	if distSourceSimple == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь:", pathSourceSimple, ", его длина:", distSourceSimple)
	}

	var startPoint, finishPoint = 0, 5
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа со смешанным ограничением для пути,")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	var pathSourceMix, distSourceMix = DeijkstraAlgorithmForAuxGraph(auxGraph, startPoint, finishPoint, 1, len(graph))
	if distSourceMix == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на вспомогательном графе:", pathSourceMix, ", его длина:", distSourceMix)
		var truePath = MakeSourcePathForMix(pathSourceMix, auxGraph, len(graph))
		fmt.Println("Путь на исходном графе:")
		fmt.Println(truePath)
	}
}

func BarrierProgramm() {
	var graphWithBarrier, barlevel = ReadGraphForBarrier("Barrier_1.txt")

	fmt.Println("Список смежности графа:")
	for i := 0; i < len(graphWithBarrier); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graphWithBarrier[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graphWithBarrier[i]); j++ {
			fmt.Println(graphWithBarrier[i][j])
		}
	}
	fmt.Println()

	//граф после удаления лишних дуг
	graphWithBarrier = DeleteExcessEdges(graphWithBarrier)
	fmt.Println("Список смежности графа после удаления лишних дуг:")
	for i := 0; i < len(graphWithBarrier); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graphWithBarrier[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graphWithBarrier[i]); j++ {
			fmt.Println(graphWithBarrier[i][j])
		}
	}
	//fmt.Println(graph)
	fmt.Println()

	var startPoint, finishPoint = 0, 7

	//fmt.Println("***************************SIMPLEWAY******************************")
	var simpleGraph = MakeSimpleGraph(graphWithBarrier)

	fmt.Println("Рассмотрим алгоритм Дейкстры для графа без ограничений для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	fmt.Println("Результат для графа без ограничений:")
	var pathSourceSimple, distSourceSimple = DeijkstraAlgorithm(simpleGraph, startPoint, finishPoint)
	if distSourceSimple == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь:", pathSourceSimple, ", его длина:", distSourceSimple)
	}
	fmt.Println()
	//fmt.Println("***************************EndSimple******************************")

	//*************************START VECTOR DEIJKSTRA**************************************
	fmt.Println("Векторный алгоритм Дейкстра для графа с барьерным ограничением:")
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	var pathSourceBarrierVecDeijkstra, distSourceBarrierVecDeijkstra = DeijkstraVectorAlgorithmForBarrier(graphWithBarrier, startPoint, finishPoint, barlevel)
	if distSourceBarrierVecDeijkstra == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на графе:", pathSourceBarrierVecDeijkstra, ", его длина:", distSourceBarrierVecDeijkstra)
	}

	fmt.Println()
	//*************************END VECTOR DEIJKSTRA**************************************

	//fmt.Println("*******************************************STARTHELPGRAPH*******************************")
	start := time.Now()
	var auxBarrierGraph = MakeAuxiliaryGraphForBarrier(graphWithBarrier, barlevel)
	//fmt.Println("Список смежности вспомогательного графа:")
	//for i := 0; i < len(auxBarrierGraph); i++ {
	//fmt.Println("Дуги, выходящие из вершины", i, ":")
	//if len(auxBarrierGraph[i]) == 0 {
	//fmt.Println("Из этой вершины дуги не исходят")
	//}
	//for j := 0; j < len(auxBarrierGraph[i]); j++ {
	//fmt.Println(auxBarrierGraph[i][j])
	//}
	//}
	//fmt.Println(auxGraph)
	//fmt.Println("*******************************************ENDHELPGRAPH*******************************")

	//алгоритм Дейкстры
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа с")
	fmt.Println("ограничением достижимости для пути,")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	fmt.Println("Результат для графа с барьерным ограничением:")
	var pathSourceBarrier, distSourceBarrier = DeijkstraAlgorithmForAuxGraph(auxBarrierGraph, startPoint, finishPoint, barlevel, len(graphWithBarrier))
	if distSourceBarrier == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на вспомогательном графе:", pathSourceBarrier, ", ")
		fmt.Println("его длина:", distSourceBarrier)
		var truePathBarrier = MakeSourcePathForBarrier(pathSourceBarrier, graphWithBarrier, auxBarrierGraph, len(graphWithBarrier))
		fmt.Println("Путь на исходном графе:")
		fmt.Println(truePathBarrier)
	}
	duration := time.Since(start)
	fmt.Println(duration)
}

func MagnetProgramm() {
	var graphWithMagnet, maglevel = ReadGraphForMagnet("MagnetGraph.txt")

	fmt.Println("Список смежности графа:")
	for i := 0; i < len(graphWithMagnet); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graphWithMagnet[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graphWithMagnet[i]); j++ {
			fmt.Println(graphWithMagnet[i][j])
		}
	}
	fmt.Println()

	//граф после удаления лишних дуг
	graphWithMagnet = DeleteExcessEdges(graphWithMagnet)
	fmt.Println("Список смежности графа после удаления лишних дуг:")
	for i := 0; i < len(graphWithMagnet); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graphWithMagnet[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graphWithMagnet[i]); j++ {
			fmt.Println(graphWithMagnet[i][j])
		}
	}
	//fmt.Println(graph)
	fmt.Println()

	//fmt.Println("*******************************************STARTHELPGRAPH*******************************")
	var auxMagnetGraph = MakeAuxiliaryGraphForMagnet(graphWithMagnet, maglevel)
	fmt.Println("Список смежности вспомогательного графа:")
	for i := 0; i < len(auxMagnetGraph); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(auxMagnetGraph[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(auxMagnetGraph[i]); j++ {
			fmt.Println(auxMagnetGraph[i][j])
		}
	}
	//fmt.Println(auxGraph)
	//fmt.Println("*******************************************ENDHELPGRAPH*******************************")

	var startPoint, finishPoint = 0, 10

	//fmt.Println("***************************SIMPLEWAY******************************")
	var simpleGraph = MakeSimpleGraph(graphWithMagnet)

	fmt.Println("Рассмотрим алгоритм Дейкстры для графа без ограничений для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	fmt.Println("Результат для графа без ограничений:")
	var pathSourceSimple, distSourceSimple = DeijkstraAlgorithm(simpleGraph, startPoint, finishPoint)
	if distSourceSimple == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь:", pathSourceSimple, ", его длина:", distSourceSimple)
	}
	fmt.Println()
	//fmt.Println("***************************EndSimple******************************")

	//*************************START VECTOR DEIJKSTRA**************************************
	fmt.Println("Векторный алгоритм Дейкстра для графа с магнитным ограничением:")
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	var pathSourceMagnetVecDeijkstra, distSourceMangetVecDeijkstra = DeijkstraVectorAlgorithmForMagnet(graphWithMagnet, startPoint, finishPoint, maglevel)
	if distSourceMangetVecDeijkstra == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на графе:", pathSourceMagnetVecDeijkstra, ", его длина:", distSourceMangetVecDeijkstra)
	}

	fmt.Println()
	//*************************END VECTOR DEIJKSTRA**************************************

	//алгоритм Дейкстры
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа с ограничением магнитности для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	fmt.Println("Результат для графа с ограничением магнитности:")
	var pathSourceBarrier, distSourceBarrier = DeijkstraAlgorithmForAuxGraph(auxMagnetGraph, startPoint, finishPoint, maglevel, len(graphWithMagnet))
	if distSourceBarrier == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на вспомогательном графе:", pathSourceBarrier, ", его длина:", distSourceBarrier)

		var truePathMagnet = MakeSourcePathForBarrier(pathSourceBarrier, graphWithMagnet, auxMagnetGraph, len(graphWithMagnet))
		fmt.Println("Путь на исходном графе:")
		fmt.Println(truePathMagnet)
	}
}

func MagnetBarrierProgramm() {
	var graphWithMagnetBarrier, maglevel = ReadGraphForMagnet("MagnetBarrierGraph_hard.txt")

	fmt.Println("Список смежности графа:")
	for i := 0; i < len(graphWithMagnetBarrier); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graphWithMagnetBarrier[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graphWithMagnetBarrier[i]); j++ {
			fmt.Println(graphWithMagnetBarrier[i][j])
		}
	}
	fmt.Println()

	graphWithMagnetBarrier = DeleteExcessEdges(graphWithMagnetBarrier)
	fmt.Println("Список смежности графа после удаления лишних дуг:")
	for i := 0; i < len(graphWithMagnetBarrier); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(graphWithMagnetBarrier[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(graphWithMagnetBarrier[i]); j++ {
			fmt.Println(graphWithMagnetBarrier[i][j])
		}
	}
	//fmt.Println(graph)
	fmt.Println()

	//fmt.Println("*******************************************STARTHELPGRAPH*******************************")
	var auxMagnetBarrierGraph = MakeAuxiliaryGraphForMagnetBarrier(graphWithMagnetBarrier, maglevel)
	fmt.Println("Список смежности вспомогательного графа:")
	for i := 0; i < len(auxMagnetBarrierGraph); i++ {
		fmt.Println("Дуги, выходящие из вершины", i, ":")
		if len(auxMagnetBarrierGraph[i]) == 0 {
			fmt.Println("Из этой вершины дуги не исходят")
		}
		for j := 0; j < len(auxMagnetBarrierGraph[i]); j++ {
			fmt.Println(auxMagnetBarrierGraph[i][j])
		}
	}
	//fmt.Println(auxGraph)
	//fmt.Println("*******************************************ENDHELPGRAPH*******************************")
	var startPoint, finishPoint = 0, 9
	//fmt.Println("***************************SIMPLEWAY******************************")
	var simpleGraph = MakeSimpleGraph(graphWithMagnetBarrier)

	fmt.Println("Рассмотрим алгоритм Дейкстры для графа без ограничений для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	fmt.Println("Результат для графа без ограничений:")
	var pathSourceSimple, distSourceSimple = DeijkstraAlgorithm(simpleGraph, startPoint, finishPoint)
	if distSourceSimple == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь:", pathSourceSimple, ", его длина:", distSourceSimple)
	}
	fmt.Println()
	//fmt.Println("***************************EndSimple******************************")

	//*************************START VECTOR DEIJKSTRA**************************************
	fmt.Println("Векторный алгоритм Дейкстра для графа с магнитным ограничением:")
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	var pathSourceMagnetVecDeijkstra, distSourceMangetVecDeijkstra = DeijkstraVectorAlgorithmForMagnetBarrier(graphWithMagnetBarrier, startPoint, finishPoint, maglevel)
	if distSourceMangetVecDeijkstra == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на графе:", pathSourceMagnetVecDeijkstra, ", его длина:", distSourceMangetVecDeijkstra)
	}

	fmt.Println()
	//*************************END VECTOR DEIJKSTRA**************************************
	//алгоритм Дейкстры
	fmt.Println("Рассмотрим алгоритм Дейкстры для графа с ограничением магнитности для пути, ")
	fmt.Println("который начинается в вершине", startPoint, "и заканчивается в вершине", finishPoint, ":")
	fmt.Println("Результат для графа с ограничением магнитности:")
	var pathSourceBarrier, distSourceBarrier = DeijkstraAlgorithmForAuxGraph(auxMagnetBarrierGraph, startPoint, finishPoint, maglevel, len(graphWithMagnetBarrier))
	if distSourceBarrier == int(^uint(0)>>1) {
		fmt.Println("Пути не существует")
	} else {
		fmt.Println("Путь на вспомогательном графе:", pathSourceBarrier, ", его длина:", distSourceBarrier)

		var truePathMagnet = MakeSourcePathForBarrier(pathSourceBarrier, graphWithMagnetBarrier, auxMagnetBarrierGraph, len(graphWithMagnetBarrier))
		fmt.Println("Путь на исходном графе:")
		fmt.Println(truePathMagnet)
	}
}

func ReadGraphForBarrierSpeedTest(filename string) ([][]trio, int) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	n, m, barlevel := 0, 0, 0
	fmt.Fscanln(f, &n, &m, &barlevel)
	graph := make([][]trio, n)
	for i := 0; i < m; i++ {
		edge := trio{}
		var vertex int
		fmt.Fscanln(f, &vertex, &edge.EndPoint, &edge.Weight, &edge.EdgeType)
		graph[vertex] = append(graph[vertex], edge)
	}
	return graph, barlevel
}

func MakeAuxBarrierAndDeijkstra(graph [][]trio, barlevel int, startPointDeijkstra int, finishPointDeijkstra int) {

	defer func(t time.Time) {
		fmt.Println("Время работы на вспомогательном графе:", time.Since(t))
	}(time.Now())

	DeijkstraAlgorithmForAuxGraph(MakeAuxiliaryGraphForBarrier(graph, barlevel), startPointDeijkstra, finishPointDeijkstra, barlevel, len(graph))
}

func BarrierSpeedTestProgramm() {
	var graph, barlevel = ReadGraphForBarrierSpeedTest("Barrier_1.txt")
	var startPoint, finishPoint = 0, 7
	MakeAuxBarrierAndDeijkstra(graph, barlevel, startPoint, finishPoint)
	_, _ = DeijkstraVectorAlgorithmForBarrier(graph, startPoint, finishPoint, barlevel)
}

func ReadGraphForMagnetSpeedTest(filename string) ([][]trio, int) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n, m, maglevel := 0, 0, 0
	fmt.Fscanln(f, &n, &m, &maglevel)
	graph := make([][]trio, n)
	for i := 0; i < m; i++ {
		edge := trio{}
		var vertex int
		fmt.Fscanln(f, &vertex, &edge.EndPoint, &edge.Weight, &edge.EdgeType)
		graph[vertex] = append(graph[vertex], edge)
	}
	return graph, maglevel
}

func MakeAuxMagnetAndDeijkstra(graph [][]trio, maglevel int, startPointDeijkstra int, finishPointDeijkstra int) {

	defer func(t time.Time) {
		fmt.Println("Время работы на вспомогательном графе:", time.Since(t))
	}(time.Now())

	DeijkstraAlgorithmForAuxGraph(MakeAuxiliaryGraphForMagnet(graph, maglevel), startPointDeijkstra, finishPointDeijkstra, maglevel, len(graph))

}

func MagnetSpeedTestProgramm() {
	var graph, maglevel = ReadGraphForMagnetSpeedTest("MagnetGraphTest.txt")
	var startPoint, finishPoint = 0, 200
	MakeAuxMagnetAndDeijkstra(graph, maglevel, startPoint, finishPoint)
	DeijkstraVectorAlgorithmForMagnet(graph, startPoint, finishPoint, maglevel)
}

func MakeAuxMagnetBarrierAndDeijkstra(graph [][]trio, maglevel int, startPointDeijkstra int, finishPointDeijkstra int) {

	defer func(t time.Time) {
		fmt.Println("Время работы на вспомогательном графе:", time.Since(t))
	}(time.Now())

	DeijkstraAlgorithmForAuxGraph(MakeAuxiliaryGraphForMagnetBarrier(graph, maglevel), startPointDeijkstra, finishPointDeijkstra, maglevel, len(graph))
}

func MagnetBarrierSpeedTestProgramm() {
	var graph, maglevel = ReadGraphForMagnetSpeedTest("MagnetBarrierGraph_hard_Test.txt")
	var startPoint, finishPoint = 0, 9
	MakeAuxMagnetBarrierAndDeijkstra(graph, maglevel, startPoint, finishPoint)
	DeijkstraVectorAlgorithmForMagnetBarrier(graph, startPoint, finishPoint, maglevel)
}

func main() {
	var typeOfLimit string

	fmt.Print("Введите тип ограничения на графе: ")
	fmt.Fscan(os.Stdin, &typeOfLimit)

	//typeOfLimit = "magbar"

	switch typeOfLimit {
	case "mix":
		MixProgramm()
	case "bar":
		BarrierProgramm()
	case "mag":
		MagnetProgramm()
	case "magbar":
		MagnetBarrierProgramm()
	case "barSpeedTest":
		BarrierSpeedTestProgramm()
	case "magSpeedTest":
		MagnetSpeedTestProgramm()
	case "magbarSpeedTest":
		MagnetBarrierSpeedTestProgramm()
	default:
		panic(fmt.Sprintf("%v", typeOfLimit))
	}

}
