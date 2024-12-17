# Start our image from a Debian stable
FROM debian:stable

# Update Debian package manager indexes
RUN apt-get update -yq

# Install htop
RUN apt-get install -yq htop

# Run HTOP by default
CMD ["/usr/bin/htop"]