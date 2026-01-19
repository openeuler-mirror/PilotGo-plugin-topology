var globals = this;

function require(m) {
    var module = {}
    var keys = Object.keys(globals)
    for (var i in keys) {
        var key = keys[i]
        module[key] = globals[key]
    }
    return module;
}

Promise.prototype.finally = function (f) {
    return this.then(function (value) {
      return Promise.resolve(f()).then(function () {
        return value;
      });
    }, function (err) {
      return Promise.resolve(f()).then(function () {
        throw err;
      });
    });
  };
