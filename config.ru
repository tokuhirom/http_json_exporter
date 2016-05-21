require 'json'

run Proc.new {
  [200, [], [JSON.generate({
    foo: 1,
    bar: {
      baz: 2
    },
		k: [ 5,9,6,3],
		"mem.free" => 5963,
    yo: nil,
    i: 'ppp',
  })]]
}
