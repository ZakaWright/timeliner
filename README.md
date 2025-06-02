# Timeliner

This is an Incident Response timeline tool made for DSU's CSC_842 Security Tool Development Course

## Project Overview/Motivation

This project was designed to end the 'spreadsheet of death' common in incident response. A few alternative tools exist such as [Aurora IR](https://github.com/cyb3rfox/Aurora-Incident-Response), but are out of support or lack modern collaboration features.

## Tech Stack
This project is built using the GOTH stack with a PostgreSQL database. The GOTH stack is made up of a Golang webserver using Templ templates and HTMX.
The GOTH stack was chosen becaue I want to become more proficient in Golang and I have heard great things about the simplicity of using Templ and HTMX.

## Current Problems
- Issue with converting times from the database

## Future Improvements
- Enable interacivity with Web Sockets for real time collaboration
- Improve Incident and Event Details layouts
- MITRE ATT&CK mappings for events
- MITRE ATT&CK view
- Make release as a Docker container
- Fix times
- Timezone choices

## Requirements
- Track incidents and events
- Real time collaboration
