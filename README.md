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
- Übersicht des Notenspiegels als Website per Webserver auf Port 1234

## Installation
##Download 
```bash
wget https://github.com/zottelchin/Notenservice/raw/master/Notenservice.zip
unzip Notenservice.zip
rm -r __MACOSX
cd Notenservice
chmod +x server
```
