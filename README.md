# Timeliner

This is an Incident Response timeline tool made for DSU's CSC_842 Security Tool Development Course

## Project Overview/Motivation

This project was designed to end the 'spreadsheet of death' common in incident response. A few alternative tools exist such as [Aurora IR](https://github.com/cyb3rfox/Aurora-Incident-Response), but are out of support or lack modern collaboration features.

## Run the project
The easiest way to run this project is with Docker.
In `docker-compose.yml` replace the TODO with a password for the database user.
To get run the project, first clone the `env-sample` file to `.env` and replace the TODO variables with your own values.
Run `docker-compose up` to start the container.

## Tech Stack
This project is built using the GOTH stack with a PostgreSQL database. The GOTH stack is made up of a Golang webserver using Templ templates and HTMX.
The GOTH stack was chosen becaue I want to become more proficient in Golang and I have heard great things about the simplicity of using Templ and HTMX.

## Future Improvements
- Expand SSE use to also work for events, endpoints, and the timeline view
- Improve Incident and Event Details layouts
- MITRE ATT&CK view
- Visual map of the incident
- Timezone choices

## New Improvements
- Made event updates interactive using Server Sent Events (SEE)
- Updated the database to include MITRE tactics
- Containerization

## Requirements
- Track incidents and events
- Real time collaboration
