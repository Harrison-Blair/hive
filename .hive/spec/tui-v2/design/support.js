// support.js renders one hive design board outside the canvas.
// A board script defines `class Component extends DCLogic` with renderVals(),
// which returns {rows: [{runs: [{text, color, weight, width}]}]}. This file
// runs that script and fills the board's <sc-for> row template with the rows.
(function () {
  function render() {
    var script = document.querySelector('script[data-dc-script]');
    var host = document.querySelector('x-dc');
    if (!script || !host) return;
    var DCLogic = function () {};
    var Component = new Function('DCLogic',
      script.textContent.replace(/class Component extends DCLogic/, 'var Component = class') + '\nreturn Component;')(DCLogic);
    var rows = new Component().renderVals().rows;
    var outer = host.querySelector('sc-for');
    var box = outer.parentElement;
    var frag = document.createDocumentFragment();
    rows.forEach(function (row) {
      var line = document.createElement('div');
      line.style.height = '17px';
      row.runs.forEach(function (run) {
        var span = document.createElement('span');
        span.style.cssText = 'display:inline-block;vertical-align:top;overflow:hidden';
        span.style.width = run.width;
        span.style.color = run.color;
        span.style.fontWeight = run.weight;
        span.textContent = run.text;
        line.appendChild(span);
      });
      frag.appendChild(line);
    });
    box.replaceChild(frag, outer);
    var helmet = host.querySelector('helmet');
    if (helmet) while (helmet.firstChild) document.head.appendChild(helmet.firstChild);
    host.replaceWith(box);
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', render);
  else render();
})();
