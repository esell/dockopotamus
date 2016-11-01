# Dockopotamus

A shitty attempt at a honeypot/sandbox that uses docker


## WARNING WARNING WARNING

**This is likely poorly implemented and a potential security nightmare**

**Currently this must run as root to interact with Docker**

**Currently users get a container with no safeguards**


## What?

For whatever reason I thought it would be cool to build a Go based honeypot to see what all these random bots and crackers are doing with their lives. I'm not a security person but I do like to learn so off I went.

The first attempt was heavily influenced by https://github.com/traetox/sshForShits and https://gist.github.com/jpillora/b480fde82bff51a06238. 
When I say "heavily influenced" I mean I basically just cut & paste their code, stripped out a few things and then called it my own. 
This was good, I liked it but I quickly realized that there was going to be a lot of work for me to actually implement some of the functionality I wanted. 
I wanted these people to get a working environment so that I could see what they were downloading, talking to, etc and the existing code only had mock/fake commands. 
Being super lazy and a poor programmer I gave up and re-thought my approach and decided that each user should get their own sandbox and what better way than giving them their own personal Docker container!?!?!


Iteration two became just a stripped down SSH server that drops users into a new Docker container that has [Snoopy](https://github.com/a2o/snoopy) installed. The container mounts a volume on the local server where the Snoopy logs are saved. 
This lets me look at what the user was doing on the system with the additional bonus that I could then kick them off and take the container somewhere else to look at. 
Of course there are some possible issues with this, the biggest is that the user has a "real" system which means they can kick of real attacks. 
As much as I enjoy being part of a global botnet, I am already stretched pretty thin with regards to free time so I don't feel that I can give the bot net my all. 
I am working on a way to limit the potential damage a user could cause from the container while still making it feel like a "real" system (see [issue #1](https://github.com/esell/dockopotamus/issues/1)). One idea is to limit the outgoing bandwidth from the container so that an attack could still technically kick off, but the amount of power given to it would be very small. 
At that point some logic can kick in to shut the container down. 

Iteration three is currently a WIP but will focus on more of the safeguards needed to make this something you are not deathly afraid to run.


# How?

You'll need [GB](https://getgb.io/) to build so install that first. Clone this repo and then do a `gb build` from inside of it, your executable will be in `bin/`.

You'll also need Docker installed and working as well as the `esell/dockopotamus` container image so go ahead and build that with the provided Dockerfile.

Currently all Snoopy logs are set to go to `/logs` so you'll need to create that and make it writeable.

Once you have all of that just run dockopotamus as root (sigh) and watch the fun roll in.

