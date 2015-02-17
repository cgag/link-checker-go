The interesting stuff is in
crawl/crawl.go.  I think the code is a bit messy, and as a result it's not
as concurrent as I'd like, but instead it's concurrent at each level of 
crawl depth. That is, every link on the first page is crawled in 
parallel, but none of the new child links are crawled until all 
the links on the first page are done. 

I wrote this mostly to experiment with Go. Coming from a functional 
background, hating on Go is something of a sport but I wanted to give it a 
fair chance.  I found it to be a mixed bag.  I really like that so much
is covered by the standard library.  Having to evaluate libraries to do every
little thing is frustrating in Haskell, it's nice to be able to just use
the standard library for almost everything.  The tools and vim integration
are fantastic as well.  Quality autoformatting on every save, and good
type-aware autocompletion.  The language itself is frustrating.  Having to
write simple maps and filters as for loops is super painful and verbose.  The
language involves a lot of typing.
