package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

type faceIndex struct {
	f1 []int32 // [v1, uv1, n1]
	f2 []int32 // [v2, uv2, n2]
	f3 []int32 // [v3, uv3, n3]
}

type objModel struct {
	string   name
	vertices []mgl32.Vec3
	uvs      []mgl32.Vec2
	normals  []mgl32.Vec3
	faces    []faceIndex
}

func readOBJ(filePath string) (objModel, error) {
	file, err := os.Open("lowPolySphere.obj")
	defer file.Close()

	if err != nil {
		log.Fatalf("failed opening obj file: %s", err)
	}

	fileScanner := bufio.NewScanner(file)
	var model objModel
	for fileScanner.Scan() {
		text := fileScanner.Text()
		values := strings.Split(text, " ")

		switch values[0] {
    case "o"
      // Mesh name
      model.name = values[1]
		case "v":
			// Vertice
			x, _ := strconv.ParseFloat(values[1], 32)
			y, _ := strconv.ParseFloat(values[2], 32)
			z, _ := strconv.ParseFloat(values[3], 32)
			model.vertices = append(model.vertices, mgl32.Vec3{float32(x), float32(y), float32(z)})
		case "vt":
			// uvs
			u, _ := strconv.ParseFloat(values[1], 32)
			v, _ := strconv.ParseFloat(values[2], 32)
			model.uvs = append(model.uvs, mgl32.Vec2{float32(u), float32(v)})
		case "vn":
			// Vertice normal
			x, _ := strconv.ParseFloat(values[1], 32)
			y, _ := strconv.ParseFloat(values[2], 32)
			z, _ := strconv.ParseFloat(values[3], 32)
			model.normals = append(model.normals, mgl32.Vec3{float32(x), float32(y), float32(z)})
		case "f":
			// face indices
			// e.g. 24/33/37 31/28/37 37/47/37
			f1text := strings.Split(values[1], "/")
			f2text := strings.Split(values[2], "/")
			f3text := strings.Split(values[3], "/")

			var face faceIndex

			f1v1, _ := strconv.ParseInt(f1text[0], 10, 32)
			f1uv1, _ := strconv.ParseInt(f1text[1], 10, 32)
			f1n1, _ := strconv.ParseInt(f1text[2], 10, 32)
			face.f1 = append(face.f1, int32(f1v1), int32(f1uv1), int32(f1n1))

			f2v1, _ := strconv.ParseInt(f2text[0], 10, 32)
			f2uv1, _ := strconv.ParseInt(f2text[1], 10, 32)
			f2n1, _ := strconv.ParseInt(f2text[2], 10, 32)
			face.f1 = append(face.f1, int32(f2v1), int32(f2uv1), int32(f2n1))

			f3v1, _ := strconv.ParseInt(f3text[0], 10, 32)
			f3uv1, _ := strconv.ParseInt(f3text[1], 10, 32)
			f3n1, _ := strconv.ParseInt(f3text[2], 10, 32)
			face.f1 = append(face.f1, int32(f3v1), int32(f3uv1), int32(f3n1))
		}

	}

	return model, nil
}
