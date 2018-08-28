# Notenservice
Notenübersicht und Mail-Benachrichtigung für das LSF der OvGU Magdeburg.

## Motivation
In der Klausurenphase wollte ich mich nicht immer mich wieder ins LSF einloggen um zu sehen ob eine weitere Klausur eingetragen wurde in den Notenspiegel. Zuerst habe ich es mit einem Plugin im Browser immer wieder neu laden lassen. Aber das war nicht so das Wahre.
Ich wollte gerne eine Benachrichtigung per Mail, wenn eine Neue Note raus kommt.

## Funktionsumfang
Mitlerweile kann das Programm zwei Sachen:
- Per Mail über neue Noten informieren 
  - mit Angabe über das Fach und die erreichte Note
  - Aktualisierungsinterval: 1h
- Übersicht des Notenspiegels als Website per Webserver auf Port 3412

## Screenshots
Die Übersicht über die Noten (_Beispieleinträge_) per Browser:
![vgy.me](https://vgy.me/SViMg1.png)

Die Mailbenachrichtigung:
![vgy.me](https://vgy.me/MyXXUW.png)
## Installation
```bash
mkdir /var/noten && cd /var/noten && mkdir frontend

wget https://raw.githubusercontent.com/zottelchin/Notenservice/master/frontend/milligram.min.css -O frontend/milligram.min.css  
wget https://raw.githubusercontent.com/zottelchin/Notenservice/master/frontend/notenuebersicht.html -O frontend/notenuebersicht.html  
wget https://raw.githubusercontent.com/zottelchin/Notenservice/master/frontend/vue.min.js -O frontend/vue.min.js
wget https://github.com/zottelchin/Notenservice/releases/download/v1.2/Notenservice  
chmod +x Notenservice

wget https://raw.githubusercontent.com/zottelchin/Notenservice/master/notenservice.service -O /etc/systemd/system/notenservice.service
wget https://raw.githubusercontent.com/zottelchin/Notenservice/master/example.config.yml -O config.yml
nano config.yml

systemctl enable notenservice
systemctl start notenservice
systemctl status notenservice

```
