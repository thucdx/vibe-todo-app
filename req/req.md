Requirements
A minimalist todo app.

# Tech stack:
- Microservices with website written with react, html, css (coulb be tailwind)
- Backend service written in golang
- DB can be in postgres
- Support distributed, at least 2 nodes of webservice with a single instance of DB
- Containerize with Docker compose

## Function
- User can add / delete / markdone a task
- There is calendar showing number of task done / total task in that day. 
- Active view is today. 
- User can move a task from a day to another day by drag with mouse, or manual change the date
- Each task is assigned 1 points by defaults, but user can change that number
- There is a chart showing number of task done for each day. Can be view by day or by week or by month. Chart can show by value of total task done, or value of points done.


