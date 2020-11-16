package main

import (
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"
)

//Kernel is an interpolator that blends source pixels weighted by a symmetric kernel function.
type Kernel struct {
	Support float64
	At      func(t float64) float64
}

//godoc.org/golang.org/x/image/draw ___ Variables
var (
	CatmullRom = &Kernel{2, func(t float64) float64 {
		if t < 1 {
			return (1.5*t-2.5)*t*t + 1
		}
		return ((-0.5*t+2.5)*t-4)*t + 2
	}}
)

func main() {

	size := flag.String("s", "800x600", "Size")
	outFolder := flag.String("o", "./", "Output folder.")
	flag.Parse()

	//find x Zeichen in der Eingabe-Satz
	i := strings.Index(*size, "x")
	if i > -1 {
		//ohne Probleme weite...
	} else {
		fmt.Println("Ungültiges Format für maximale Größe. Format ist zB -s 800x600")
		os.Exit(1)
	}

	//Split-Function (RUF)
	Breite, Höhe := Split(*size, "x")

	//flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, os.ErrInvalid)
		return
	}

	// Create Folder
	if _, err := os.Stat(*outFolder); os.IsNotExist(err) {
		os.Mkdir(*outFolder, 0755)
	}

	//ResizeImage-Function (RUF)
	for _, f := range flag.Args() {
		ResizeImage(Breite, Höhe, f, *outFolder)
	}

}
func Split(size, x string) (int, int) {
	var Breite, Höhe int
	tmp := strings.Split(size, x)
	values := make([]int, 0, len(tmp))
	for _, raw := range tmp {
		v, err := strconv.Atoi(raw)
		if err != nil {
			log.Print(err)
			continue
		}
		values = append(values, v)
	}
	Breite = values[0]
	Höhe = values[1]
	fmt.Println("W=", Breite, "H=", Höhe)
	return Breite, Höhe
}

func ResizeImage(w, h int, file, destFolder string) error {

	src, err := os.Open(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "leider können wir die gewünschte datei: "+file+" nicht finden.")
		return nil
	}
	defer src.Close()

	//decode image
	imgSrc, t, err := image.Decode(src)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Die Datei: "+file+" ist nicht einem Bild.", "Bitte laden Sie nur das Foto im JPG,GIF,PNG-Format")
		return nil
	}
	fmt.Println("Type of image:", t, ", Name of image:", file)
	//rectange of image

	rctSrc := imgSrc.Bounds()
	fmt.Println("Original Width:", rctSrc.Dx())
	fmt.Println("Original Height:", rctSrc.Dy())
	//AspectRatio-Function (RUF)
	x, y := AspectRatio(w, h, rctSrc.Dx(), rctSrc.Dy())

	fmt.Print("w:", x, ", h:", y, "\n")

	imgDst := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.CatmullRom.Scale(imgDst, imgDst.Bounds(), imgSrc, rctSrc, draw.Over, nil)

	hash := sha256.New()
	hash.Write([]byte(file))
	md := hash.Sum(nil)
	NewFile := base32.StdEncoding.EncodeToString(md)
	fmt.Println(file, "->", NewFile)
	fmt.Println("___________________________")

	dst, err := os.Create(destFolder + "/" + NewFile + ".jpg")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Das System kann den angegebenen Pfad für: "+file+" nicht finden.")
		return err
	}
	defer dst.Close()

	//encode resized image with switch-case
	switch t {
	case "jpeg":
		if err := jpeg.Encode(dst, imgDst, &jpeg.Options{Quality: 100}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	case "gif":
		if err := gif.Encode(dst, imgDst, nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	case "png":
		if err := png.Encode(dst, imgDst); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	default:
		fmt.Fprintln(os.Stderr, "format error")
	}

	return nil
} // end of func ResizeImage

func AspectRatio(w, h, rctSrcDx, rctSrcDy int) (int, int) {

	// Querformat     if width>height
	x := w
	y := (w * rctSrcDy) / rctSrcDx
	if y > h {
		// anpassen y (Höhe)
		y = h
		x = (h * rctSrcDx) / rctSrcDy
	}
	// Hochformat   if width<height
	if rctSrcDx < rctSrcDy {
		y = h
		x = (h * rctSrcDx) / rctSrcDy
		if x > w {
			// anpassen x (Breit)
			x = w
			y = (w * rctSrcDy) / rctSrcDx
		}
	}
	return x, y
}
