---
css:
  - Презентации/common.css
highlightTheme: nord
---
<link rel="stylesheet"          href="https://fonts.googleapis.com/css2?family=Noto+Sans">
<style>
.nohljsln .hljs-ln-numbers  {
display: none;
}
</style>

## Контейнеры. Ликбез

Здравствуйте, я ваш Docker

![[Pasted image 20230927205036.png]]

note: Привет, народ. Сегодня мы познакомимся с контейнеризацией, узнаем как она может упрощать жизнь, кто такой этот Docker и почему "На моей машине все работает" - больше не отговорка.

---

## Вначале были требования

![[giphy.gif]]

note: Мы любим решать задачи несколькими разными способами. Запуск приложений - не исключение. За время существования современных компьютеров мы уже несколько раз решили эту (как оказалось нетривиальную) задачу

Установка на устройство - хороший и естественный способ развертки приложений. Но любой, кто ковырялся в Arch Linux скажет вам, что разрешение конфликтов зависимостей - боль

Виртуальные машины - отличное решение. Ограничим  спектр задач "устройства", сможем его по желанию останавливать, перезапускать ,клонировать. Кайф. Но не очень эффективно.

А что же контейнеры?

---

## Контейнеры против виртуалок

![[Pasted image 20231011164903.png]] <!-- element class="fragment" -->

note: Контейнеры очень похожи на виртуалки. Они пытаются ответить на вопрос: Что если виртуалка, но жирно чуть-чуть поменьше?

Контейнеры не виртуализируют ядро ОС и железо машины. Из-за этого они намного быстрее поднимаются и меньше весят. Но при этом нельзя запустить на windows контейнер, сделанный для linux. Хотя на практике это ограничение мало где имеет значение, ведь чаще всего контейнеры работают на серверах, где и так используется linux (правда, не все так просто с облаком)

---

## Анатомия контейнера

```dockerfile[|1|3|5|7|9|10]
FROM golang:1.20.8-alpine

WORKDIR /app

COPY . .

RUN go build -o build/

EXPOSE 8080
CMD [ "./build/docker-intro" ]
``` 
<!-- element class="fragment nohljsln" -->

note: Ладно, контейнеры прикольные и хочется уже их пощупать, но для начала поймем откуда они берутся. Контейнер - это исполняемая копия образа, неизменяемой инструкции по настройке среды (думаем про классы и объекты, да). 

Создавая образ, вы берете за основу существующий (например, базовую версию какого-нибудь дистрибутива linux) и указываете что в ней нужно изменить, чтобы в конце концов запустить ваше приложение. Эти инструкции записываются в специальный файл - Dockerfile (или, если вам больше по душе OCI-стандарты, Containerfile). Слова "докерфайл" и "контейнерфайл" взаимозаменяемы

Каждая инструкция, которую вы указываете в докерфайле наслаивается на предшествующий образ. Так операции кешируются и при написании хорошего докерфайла важно двигаться от общего к частному. Инструкции, которые меняются реже, надо ставить выше

---

## Реальный пример. Сервис на Go

```shell
$ docker build --tag docker-intro \
--file Containerfile \
.
```
 <!-- element class="fragment" -->
 
```shell
$ docker run --interactive \
--tty \
docker-intro
```
 <!-- element class="fragment" -->

note: С теорией все, погнали к практике. Наша задача - развернуть на устройстве сервис, написанный на Go. Не знаю как у вас, но на моей машине Go не установлен. Но зато в репозитории с сервисом есть докерфайл. (Оговорка - в реальной среде, вероятно, был бы уже подготовлен образ с сервисом, но так бывает не всегда, особенно, когда мы говорим о продуктовых сервисах внутри компаний)

Мы можем сами собрать контейнер на основе докерфайла, используя docker build.

(После запуска видим сообщение от сервиса в консоль)

---

## Пробрасываем порты

```shell[|3-4]
$ docker run --interactive \
--tty \
--rm \
--publish 8080:8080 \
docker-intro
```
<!-- element class="fragment nohljsln" -->

note: С нашим сервисом есть небольшая проблема. Он открывает HTTP API, но у нас нет возможности им воспользоваться. Потому что контейнер изолирован, нет разницы, что он слушает на каком-то порту. Наша машина этого не делает.

Хорошо, что это легко исправить. Добавив опцию --publish при запуске контейнера мы можем связать порт на нашей машине с портом внутри контейнера, порты при этом не должны быть одинаковыми.

(После проброса видим веб-интерфейс)

---
## Учимся масштабировать

```shell[1-7|4-5]
$ docker run --interactive \
--tty \
--publish \
--restart unless-stopped \
--name docker-intro-container \
docker-intro
```
<!-- element class="fragment  nohljsln"  -->

```shell
$ docker exec --interactive --tty docker-intro-container
```
<!-- element class="fragment"  -->

note: Одна из самых сильных сторон контейнеризированных приложений - возможность легкого масштабирования и перезапуска. Для этого можно использовать опцию --restart.

Заметили странное? Похоже, что наш сервис не запоминает информацию между перезапусками. Для некоторых приложений такое поведение годится, но в нашем случае выглядит как баг.

Мы можем "влезть" внутрь контейнера, при помощи команды docker exec --interactive --tty. Как видите, роль постоянной памяти для нашего сервиса играет файл в файловой системе. Контейнеры могут изменять свою файловую систему, но после остановки, все изменения пропадают. Каждый новый запущенный контейнер всегда стартует с той точки, что определена в образе, на основе которого он построен.

---

## Постоянная память

```shell
$ docker volume create docker-intro-volume
```
<!-- element class="fragment" style="width: 100%"  -->

```shell[|6-7]
$ docker run --interactive \
--tty \
--publish \
--restart unless-stopped \
--name docker-intro-container \
--mount \
type=volume,source=docker-intro-volume,target=/app/data \
docker-intro
```
<!-- element class="fragment" style="width: 100%"  -->

note: Чтобы исправить этот баг мы можем смонтировать том внутрь контейнера. Для этого надо создать том и смонтировать его, используя опцию --mount.

Другой способ предоставить контейнеру постоянную память - привязать директорию с локальной машины к файловой системе в контейнере. Этим пользуются, например, при разработке изнутри контейнера

(Добавляем том, наблюдаем постоянную память после перезапусков)

---

## Конец

(на самом деле нет) <!-- element class="fragment" -->

::: block <!-- element class="r-stack" -->

::: block <!-- element class="fragment current-visible" -->
- [docker-compose](https://docs.docker.com/compose/)
- [контейнеризированные среды разработки](https://docs.docker.com/desktop/dev-environments/)
- [уменьшение размеров образов](https://blog.codacy.com/five-ways-to-slim-your-docker-images/)
- [многоступенчатые сборки](https://docs.docker.com/build/building/multi-stage/)
:::

![[giphy (1).gif]]
<!-- element class="fragment" -->
:::

note: Вот в общем-то и все, что нужно знать о контейнерах, чтобы не бояться их и начать их использовать. Я призываю вас экспериментировать, читать доки, контейнеризировать свои сервисы и осваивать этот офигенский инструмент.

Вот несколько интересных вещей, которые можно изучить слету:
- docker-compose для запуска нескольких связанных сервисов. Очень полезно, чтобы не мучаться с разными базами и стаками для разных проектов
- контейнеризированные среды разработки. Все, что нужно, чтобы начать работать над проектом внутри одного контейнера. Супер-быстрый онбординг и легкая изоляция зависимостей
- как можно снизить размер образов, дабы не деплоить в прод 300+ мегабайт ради простого приложения
- многоступенчатые сборки - способ собирать сложные системы внутрь одного образа с минимумом усилий (пример - веб-приложения + сервер для их раздачи)

А на сегодня все, детишки, если у вас остались вопросы, то сейчас самое время их задать =)