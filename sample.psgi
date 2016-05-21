use JSON::PP;

my $json = encode_json({
    foo => 1,
    bar => {
        baz => 2
    },
    yo => undef,
    i => 'ppp',
});

sub { [200, [], [$json]] }
