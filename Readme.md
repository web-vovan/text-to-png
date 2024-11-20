## Описание

Пакет предназначен для конвертации любой последовательности байт в png изображение.

Флаги для настройки:

```sh
-file string
    Путь к файлу с данными
-output string
    Название файла с изображением (default "result")
```

Пример конвертации романа Булгакова "Мастер и Маргарита".
```sh
go run main.go -file bulgakov.txt -output bulgalov 
```

На выходе получаем такую картинку.

![](https://github.com/web-vovan/text-to-png/blob/main/img/bulgalov.png?raw=true)
