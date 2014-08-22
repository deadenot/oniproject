require('insert-css')(require('./app.css'))

var Vue = require('vue')

new Vue({
	el: '#app',
	components: {
		a: require('./component-a'),
		b: require('./component-b'),
		Animations: require('./animations'),
		Tools: require('./tools'),
		Tree: require('./tree'),
	},
	// require html enabled by the partialify transform
	template: require('./app.html'),
	data: {
		title: 'Hello Browserify & Vue.js!'
	}
})
