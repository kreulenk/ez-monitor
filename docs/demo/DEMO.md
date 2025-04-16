# Demo

The files contained in this directory all have to do with the generation of the demo
gif that exists within the readme of this repository.

The demo gif is generated using the VHS gif recording tool which will need to be installed on your
system to generate the gif yourself. More information on VHS can be found here -- https://github.com/charmbracelet/vhs

## Setting Up DB for Demo gif generation

For the keys defined in this demo gif to make sense, you will need to define a hosts file called `inventory.ini` in
the root of this repository. That inventory file will need three hosts in it.

# Creating Demo Gif

Once you have the proper inventory in place, run `make demo-gif` from the root of this repository.

You should now have a newly generated demo.gif!