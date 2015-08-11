<div class="panel panel-default">
  <div class="panel-heading">
    ## Problem 4.1
  </div>
  <div class="panel-body">
Calculate the velocity of an artificial satellite orbiting the Earth in a
circular orbit at an altitude of 200 km above the Earth's surface.

The radius of the earth is 6,378.14 km.

```javascript
var GM = 3.986005e14; function v(r) { return Math.sqrt(GM/r); }
var R*earth = 6378.14 * 1000; // meters
var altitude*satellite = 200 * 1000; // meters
var R*satellite = R*earth + altitude*satellite; v(R*satellite) / 1000 + " km/s";
```

where $foo = bar$
  </div>
</div> 