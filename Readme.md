
# Основная информация

Этот проект представляет из себя решение финального задания второго спринта по Go от Яндекс.Лицея - сервер с параллельным вычислителем. 

Проект состоит из двух частей: основной сервер (далее - "Оркестратор") и распределённый вычислитель (далее - "Агент"). 




## Структура запросов и ответов

"Оркестратор" имеет следующие эндпоинты:

    1) /api/v1/calculate - отвечает за непосредственное взаимодействие с пользователем,
    именно сюда пользователь направляет JSON-запрос вида:
    '{"expression": <строка с выражение>}'
    Возможные коды ответа:
    201 - выражение принято для вычисления, 
    422 - невалидные данные, 
    500 - что-то пошло не так
    --------------------------------
    Тело ответа:
    {"id": <уникальный идентификатор выражения}


    2) /api/v1/expressions - получение списка выражений
    Тело ответа:
    {
    "expressions": [
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        },
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
    ]
    }
    --------------------------------------
    Возможные коды ответа:
    200 - успешно получен список выражений
    500 - что-то пошло не так


    3) /api/v1/expressions/:id - Получение выражения по его идентификатору
    Тело ответа:
    {
    "expression":
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
    }
    ----------------------------------
    Возможные коды ответа:
    200 - успешно получено выражение
    404 - нет такого выражения
    500 - что-то пошло не так


    4)/internal/task - Получение задачи для выполнения (для "Агента")
    Тело ответа:
    {
    "task":
        {
            "id": <идентификатор задачи>,
            "arg1": <имя первого аргумента>,
            "arg2": <имя второго аргумента>,
            "operation": <операция>,
            "operation_time": <время выполнения операции>
        }
    }
    ------------------------
    Возможные коды ответа:
    200 - успешно получена задача
    404 - нет задачи
    500 - что-то пошло не так

## Принцип работы

    Если картинка не показывается как нужно, нажмите на ссылку, пожалуйста.

![alt text](https://github.com/aoktiv/calculator_project/tree/main/source/scheme.png)


## Как запустить

    1) Склонируйте репозиторий к себе на компьютер:
    git clone https://github.com/aoktiv/calculator_project

    2) Перейдите в директорию проекта, зайдите в терминал и напишите команду:
    go run main.go

    Убедитесь, что у вас установлена последняя версия golang и все 
    необходимые для работы программы пакеты!


## CURLs
    
    Код 201:

    1) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "2+2*2"
    }'
    2) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "(20*2)+5"
    }'
    3) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "17*9-10"
    }'
    4) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "(12/3)*5"
    }'

    ----------------------------------------------------

    Код 422:

    1) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "2+2*"
    }'
    2) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "(20*)+5"
    }'
    3) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "*9-10"
    }'
    4) curl --location 'localhost/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{
    "expression": "(12/3*5"
    }'
