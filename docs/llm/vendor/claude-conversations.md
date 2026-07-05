00
21
39

something isn't quite right with the passkey set up 
> The requested RPID did not match the origin or related origins.
also I think we should add a rate limit on sign ups from the same ip address
unlike the 0.5rps, burst 5, this one should be a little more strict 
only allow like one new sign up every five minute from the same IP address? 
only allow one ne sign up every minute globally? 
maybe we can add both these rules and make them configurable with our .env.example? 
Also, once passkeys are working correctly, 
I would like you to add another page available on the website 
that properly explains the concepts of passkeys in general 
as well as how we have implemented it on this site 
and how users can use this to become familiar with passkeys 
and use this website as a sandbox to learn about using passkeys 
with their iPhone or Android device. 
It isn't possible to actually sign up with just passkeys, is it? 
How do passkeys work? 
I mean what do users need to know about how passkeys work? 
Lets explain everything in plain English. 
also, can we publicly expose our csp report on the website? 
or does it not make sense? 
it is just a learning sandbox so I think it should be safe 
also we don't actually render user generated text 
like if a user were to type
```javascript 
<script>alert()</script>
```
we don't actually run this, right? we simply show this as text content? 
what about storing it on the website? Is that a problem? 
If so, we should store it properly. 
so yeah, we have pass keys support, TOTP support, and we support long, complex passwords 
so these are good practices 
the host_name::str on uptrace says `<empty string>`
not sure what that is about. should it say `virginia` when running from virginia? 
how would it get this information? 
also on the notes page, when I type this
```
O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V{g[SM@x/`O6yd"uU]NslFGb.VCh4.UnIBz2z6F7bbY6.@6UgB{#B"A5Y`vTo_V🤔🫣🦕🤧😂💦w
```
the counter says 494 / 500
however I can not type any more text 
if I try to add more text in the address bar and paste it, only up to the w shows up. 
also it would be nice to have some kind of a multiple select dropdown on the notes page so you can select whose notes you want to see 
and have your preference persisted locally and on the server so it survives a page reload 
also it would be nice to have both light mode and dark mode on the website or actually multiple themes like solarized light, solarided dark, and so on 
there are so many things we can do but we should also make sure we have proper test coverage for all the things we do 
please give me full files as well as file paths for all files that change. 
