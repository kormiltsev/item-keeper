# MVP

## 1 этап

### Клиент

Устанавливать соединение с сервером по gRPC. 

Регистрация нового клиента и авторизация существующего (парой логин-пароль). 

Получение от сервера списка всех личных данных пользователя. Вывод на экран списка, поиск и отображение конкретного объекта.

Создание нового, изменение и удаление конкретного объекта. Отправка обновленной информации на сервер.

### Сервер

При верной паре логин-пароль (или при регистрации нового пользователя) отправлять клиенту уникальный токен с ограниченным сроком действия.

Выгрузить данные из БД по клиенту и отправить в сторону агента. 

Принимать и обновлять в БД изменения. 

## 2 этап

### Клиент

Поиск по объектам по последовательности символов. Поиск по названию объекта и по полям параметров. Поиск в оперативке клиента.

Работа с данными при помощи терминала (TUI)

Добавить возможность шифровать данные (опционально для клиента). Шифрование на стороне клиента по его секретному слову. При каждом открытии приложения вводить код.

При остановке приложения сохранять данные локально. При запуске сравнивать дату последнкй синхронизации с сервером. Приоритет - сервер.

Проверять дату последнего обновления при каждой отправке обновления. Если есть расхождения, загружать новую информацию с сервера.

### Сервер

Хранить дату последнего изменения данных по клиенту.

## 3 этап

### Клиент

Загружать на сервер бинарные файлы (изображения) с привязкой к конкретному объекту.

Получать от сервера картинки и складировать их в файловой стистеме клиента. Должна сохраниться привязка с объектом.

### Сервер

Принимать и отдавать изображения. Хранить в файловом сервере, адреса в БД.

## 4 этап

### Клиент

В качестве клиента реализовать HTTP сервер. Поддержка API. Хранить на стороне клиента в map данные пользователей, которые прошли авторизацию через сервер. 
Поиск по объектам релизовать тут.

Расшифровка чувствительных данных тоже происходит тут.

### Сервер

Обеспечить балансироваку нагрузки. Воркер обслуживает очередь передачи данных в сторону агента (http server). 
Предусмотреть случай одновременного подключения большого количества клиентов по API. В приоритете отдавать данные пользователей, вторично отдача картинок.
Прерывать отдачу картинок при поступлении запроса данных на нового пользователя. 

### Дополнительный сервис по сжатию картинок

Принимает картинки, сжимает до минимально приемлемого для обложки качесвта (120х120) и вторую копию до 1 мб. Организовать очередь.

## 5 этап 

### Клиент

Прописать инструкции, описание, и прочее.

Добавить телеграм бота в качестве агента.

Добавить веб сервер в качестве агента.

Завернуть серверных клиентов в контейнеры.

### Сервер

Прописать инструкции и описание.

Завернуть в контейнер.
