#!/bin/bash

project_path=$(pwd)

# Переход в папку backend и запуск бинарника main
cd backend/
./main &

# Ожидание некоторого времени (дайте бинарнику время для запуска)
sleep 1

# Открытие нового терминала и запуск npm run serve в папке frontend/latest_news
osascript -e "tell app \"Terminal\" to do script \"cd $project_path/frontend/latest_news && sudo npm run serve\""
