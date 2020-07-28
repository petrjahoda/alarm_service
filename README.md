# Alarm Service

## Description
Go service that send alarm emails, when sql returns a row

## Additional information
* in alarms table user "," or ";" as a delimiter for email recipients
* result of sql query has to be in one column named result


## Examples
* Sends and email at the beginning of 15 hour on Saturday
```sql
select to_char(now(), 'Day') like '%Saturday%' and extract(hour from now()) = 12 as result
```
* Sends and email when more than 5 workplaces are in production
```sql
select (select count(*) from zapsi3.public.state_records where date_time_end is null and state_id=3) > 5 as result;
```    
Petr Jahoda © 2020
