# Latest News
This is a news look-up web application. It was originally made for a chatbot product that was a real-world business project.

## Functions

This application uses **Go** as the primary language and crawls webpages with **goquery**. It also used the **gg** library for generating thumbnails.

When the client makes a request without providing any keyword, the API will refer to locally stored news headlines and return a HTML5 page containing the news.

When the client provides a keyword, the API will look up the keyword at Baidu News and take note of all search results that are relevant. It will then generate a thumbnail and return a HTML5 page containing the news.

## Details

Crawling results are stored locally to avoid over-crawling.

- When the server is up, the background crawler will start working.
  - The background crawler crawls news headlines periodcially;
  - When the client makes a request without providing keyword, the news headlines are returned without any crawling.
- There are three APIs:
  - an API for returning text messages
  - an API for returning message cards
  - an HTML5 API for returning HTML5 news pages
- The chatbot should send a request to the first API and get the text response and card id. From that, it can use the card id to request a message card from the second API. The third API is used in the meanwhile to generate HTML5 pages.
- Frequency settings:
  - crawler without keyword: crawls every 12 hours
  - crawler with keyword: keeps results for each keyword in local for 1 hour
  - thumbnails without keyword: not deleted (at most 366 pcs)
  - thumbnails with keyword: uploaded to cloud servers and not kept in local
  - message cards: kept in local for 1 minute

## Code Structure

./service/
  serve.go                        server file
  news_query_handle.go            API for returning text messages
  news_card_handle.go             API for returning message cards
  news_page_handle.go             API for returning HTML5 pages
  news_crawlers.go                crawlers
  news_utils.go                   Link class, Card class, and IO helpers
  response.go                     definitions of return types
./static/                         static files
./data/                           non-static files
./templates/                      template for HTML5 pages
./main.go                         main program

## 1. Format 

Single-round task-based chatbot

## 2. Sources

Baidu News & Baidu News Search

http://news.baidu.com/

https://www.baidu.com/s?tn=news&rtt=4&bsst=1&cl=2&wd=NLP&medium=1

## 3. Triggers

- news / headline news
- news about *keyword* / *keyword* news

> Examples:
>
> news
>
> news about NLP
>
> Shanghai news
>
> ...

## 4. Bot Response

### 4.1 When keyword is not provided

#### 4.1.1 Reply Format

- Message card:

  - Title: Latest news

  - Subtitle: Here is the latest news.

  - Cover: refer to blueprint

  - Link: link to HTML5 news page

- The news page should include:

  - Current date

  - News titles

    > No changes to be made on news links.

#### 4.1.2 Sources

Choose 10 news randomly from the homepage of news.baidu.com.

### 4.2 When keyword is provided

#### 4.2.1 Reply Format

- Message card:

  - Title: Latest news on *keyword*

  - Subtitle: Here is the latest news on *keyword*.

  - Cover: refer to blueprint

  - Link: link to HTML5 news page

- The news page should include:

  - Current date
  
  - News titles

    > No changes to be made on news links.

#### 4.2.2 Sources

Use Baidu News Search with the setting "sort by time" and "media websites only". Take the first 10 search results that are revelant, i.e. the keyword is included in the title.

## Local Test

```bash
./run.sh
```
