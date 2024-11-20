package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math"
	"os"
	"flag"
)

// Сигнатура png файла
var signature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// Структура чанка
type Chunk struct {
	Lenght uint32
	Name string
	Data []byte
	Crc uint32
}

func main() {
	file := flag.String("file", "", "Путь к файлу с данными")
	output := flag.String("output", "result", "Название файла с изображением")

	flag.Parse()

	// Читаем файл
	rawBytes, err := os.ReadFile(*file)

	if err != nil {
		fmt.Println("Файл не найден:", *file)
		return
	}

	pngFileName := *output + ".png"

	// Размеры изображения
	width := calculateSize(&rawBytes)
	height := width

	// Создаем буффер байт 
	var byffer bytes.Buffer

	// Пишем сигнатуру файла
	byffer.Write(signature)

	// IHDR чанк
	ihdrChunk := createIHDRChunk(width, height)
	writeChunkToBuffer(&byffer, &ihdrChunk)

	// IDAT чанк
	idatChunk := createIDATChunk(width, height, &rawBytes)
	writeChunkToBuffer(&byffer, &idatChunk)

	// IEND чанк
	iendChunk := createIENDChunk()
	writeChunkToBuffer(&byffer, &iendChunk)

	// Пишем данные в файл с изображением
	writeBytesToFile(pngFileName, byffer.Bytes())
}

// Вычисление размера изображения
func calculateSize(bytes *[]byte) int {
	countPixels := math.Ceil(float64(len(*bytes))/3)

	return int(math.Ceil(math.Sqrt(countPixels)))
}

// Запись чанка в буфер
func writeChunkToBuffer(buffer *bytes.Buffer, chunk *Chunk) {
	// длина чанка
	binary.Write(buffer, binary.BigEndian, chunk.Lenght)
	// имя чанка
	buffer.WriteString(chunk.Name)
	// данные чанка
	buffer.Write(chunk.Data)
	// crc23 чанка
	binary.Write(buffer, binary.BigEndian, chunk.Crc)
}

// Создание IDAT чанка
func createIDATChunk(width int, height int, rawBytes *[]byte) Chunk {
	lenRawBytes := len(*rawBytes)

	// На 1 пиксель 3 байта + 1 байт фильтрации в начале каждой строки
	dataWidth := 3*width + 1

	// Срез с общим количеством байт
	data := make([]byte, dataWidth * height)

	// Пропуск вычисления пикселя
	skipPixel := false

	for i := 0; i < height; i++ {
		// В начале каждой строки устанавливаем байт фильтрации
		data[i*dataWidth] = 0

		// Пишем все байты в строке
		for j := 0; j < width; j++ {
			// Смещение для данных в чанке
			dataOffset := i*dataWidth + 3*j + 1

			// Смещение для исходных байтов
			rawBytesOffset := 3*(i*width + j)

			// Байты для 1 пикселя
			var r, g, b byte

			if (!skipPixel) {
				if rawBytesOffset < lenRawBytes {
					r = (*rawBytes)[rawBytesOffset]
				}

				if rawBytesOffset+1 < lenRawBytes {
					g = (*rawBytes)[rawBytesOffset+1]
				}

				if rawBytesOffset+2 < lenRawBytes {
					b = (*rawBytes)[rawBytesOffset+2]
				} else {
					skipPixel = true
				}
			}

			data[dataOffset] = r // R
			data[dataOffset+1] = g // G
			data[dataOffset+2] = b // B
		}
	}

	var compressedData bytes.Buffer

	zw := zlib.NewWriter(&compressedData)
	zw.Write(data)

	zw.Close()

	idatChunk := Chunk{
		Lenght: uint32(len(compressedData.Bytes())),
		Name: "IDAT",
		Data: compressedData.Bytes(),
	}

	idatChunk.Crc = calculateCrc32(&idatChunk)

	return idatChunk
}

// Создание IHDR чанка
func createIHDRChunk(width, height int) Chunk {
	var ihdrData = make([]byte, 13)

	// Ширина
	width32 := uint32(width)
	ihdrData[0] = byte(width32>>24)
	ihdrData[1] = byte(width32>>16)
	ihdrData[2] = byte(width32>>8)
	ihdrData[3] = byte(width32)

	// Высота
	height32 := uint32(height)
	ihdrData[4] = byte(height32>>24)
	ihdrData[5] = byte(height32>>16)
	ihdrData[6] = byte(height32>>8)
	ihdrData[7] = byte(height32)

	ihdrData[8] = 8 // Глубина цвета (8 бит на канал)
	ihdrData[9] = 2 // Тип цвета (2 = Truecolor, RGB)
	ihdrData[10] = 0 // Метод сжатия
	ihdrData[11] = 0 // Метод фильтрации
	ihdrData[12] = 0 // Интерлейсинг

	chunk := Chunk{
		Lenght: uint32(13),
		Name: "IHDR",
		Data: ihdrData,
	}

	chunk.Crc = calculateCrc32(&chunk)

	return chunk
}

// Создание IEND чанка
func createIENDChunk() Chunk {
	chunk := Chunk{
		Lenght: uint32(0),
		Name: "IEND",
	}

	chunk.Crc = calculateCrc32(&chunk)

	return chunk
}

// Запись байт в файл
func writeBytesToFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)

	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}

	defer file.Close()

	_, err = file.Write(data)

	if err != nil {
		return fmt.Errorf("не удалось записать данные в файл: %w", err)
	}

	return nil
}

// Чтение байт из файла
func readBytesFromFile(file string) error {
	data, err := os.ReadFile(file)

	if err != nil {
		return fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	for i, b := range data {
		fmt.Printf("%02x ", b)

		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}

	return nil
}

// Расчет crc32 чанка
func calculateCrc32(chunk *Chunk) uint32 {
	crc := crc32.NewIEEE()

	crc.Write([]byte(chunk.Name))
	crc.Write(chunk.Data)

	return crc.Sum32()
}
