# Alarm Service


## Installation
* use docker image from https://cloud.docker.com/repository/docker/petrjahoda/zapsi_service
* use linux, mac or windows version and make it run like a service

## Description
Go service that send alarm emails, when sql returns a row

## Additional information
* in alarms table user "," or ";" as a delimiter for email recipients


## Examples
* Sends and email at the beginning of 15 hour on Saturday
```sql
select where (to_char(now(), 'Day')) like '%Saturday%' and extract(hour from now()) = 15;
```
* Sends and email when more than 5 workplaces are in production
```sql
select where (select count(*) from zapsi4.public.state_records where date_time_end is null and state_id=1) > 5;
```    
www.zapsi.eu © 2020
