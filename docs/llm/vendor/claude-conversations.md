00
54
19

I have added a new /captcha page 
however, right now it is frontend only
as a demonstration of the website's ability to handle user information
we need to store this in the backend on our database
but also we need to add opentelemetry here as well 
also we need to have a leaderboard, but only for people who are signed in
on the same page, collapsed by default 
but we should remember the user setting 
if they expand it / collapse it again

also the index.html is a little strange
it says 
Get Started 
Sign up / Log in 
even though I am logged in 
and the settings page correctly displays the logout button 
and even the index / home page clearly identifies 
earlier in the page that 
I am `signed in as x`

another big feature we can add is a new page where everyone can add a new short note
something like a microblog or a tweet, plain text only, no attachments allowed. 
and everyone can see everyone else's posts. 
there are no images, no image previews, no link previews, 
and even the links in posts are not actually clickable 
however, it should be very easy to copy paste any post 
also any user should be able to delete an old post but NOT edit them
you should not be able to delete any one else's posts 
this page should also be responsive and mobile friendly 
see what I changed with the /activity page for inspiration (basically cards)
also this page should somehow autorefresh 
and this might be the most difficult part of the whole operation

please document all architectural decisions 
and please update all readme as necessary as we continue to improve this application 

please return full files as well as full path for all files that need to change
