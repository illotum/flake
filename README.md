Flake is a reasonably effective unique ID generator for high-throuput services.

Flake v1 is 12 byte long, time sortable and is comprised of:
```
 0  3    7       15       23       31
+----+----+--------+--------+--------+
|0001|                               |
+----+  Unix Time in usec   +--------+
|                           |Overflow|
+---------+--------+--------+--------+
|        Worker ID          |  CRC8  |
+---------+--------+--------+--------+
```
