# AVITO-API
**Тестовое задание для Golang стажера Backend. Деменев Юрий Михайлович **

Для того чтобы поднять проект необходимо следующее:
1) PostgreSQL работающий на порту 5432
2) Redis работающий на  хосту localhost:6379

В PostgreSQL необходимо создать базу данных "Banners".
Чтобы создать таблицы необходимо воспользоваться кодом из файла init_up.sql по пути /schema

При выполнении задания я провёл полное интеграционное тестирование. тесты находсятся в banners_test.go
по пути pkg/handler.

Я также провёл нагрузочное тестирование с помощью jmeter. Результат тестирования в папке jmeterTest/Hard_Test1/index.html
(Именно из-за этой папки github и пишет,что больше всего кода на Javasctipt и HTML)

В ходе тестирования у меня получилось добиться того что все 10000 запросов проходят успешно. В случаи увеличения числа 
запросов в секунду, определенный процент перестают проходить потому что получают connection refused.

