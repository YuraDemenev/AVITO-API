# AVITO-API
Тестовое задание для стажера Backend. Деменев Юрий Михайлович

Для того чтобы поднять проект необходимо следующее:
1) PostgreSQL работающий на порту 5432
2) Redis работающий на  хосту localhost:6379
В PostgreSQL необходимо создать базу данных "Banners".
Чтобы создать таблицы необходимо воспользоваться кодом из файла init_up.sql по пути /schema

При выполнении задания я провёл полное интеграционное тестирование. тесты находсятся в banners_test.go
по пути pkg/handler.
Я также провёл нагрузочное тестирование с помощью jmeter. Результат тестирования в папке jmeterTest

Вопросы/Проблемы с которыми я столкнулся.
Самая большая для меня проблема с которой я столкнулся, это создать sql запрос, который получая feature_id,
tag_ids вывел мне бы баннер, который имеет данную фичу и тэги. Проблема в том что в sql есть in (1,2,3),
но этот оператор работает как или 1 или 2 или 3. А мне надо было одназночно найти баннер с определеными тэгами.

Также неожиданная проблема возникла, когда я проводил полное интеграционное тестирование, пути /banner. Так как 
при сравнение того что я получил с тем что ожидаю, оказалось, что у меня расходиться время создания, так как в 
процессе разработке,я пересоздавал таблицы. Пришлось специально для этого случая конвертировать результат в struct
менять в нём время и снова конвертировать в строку.
