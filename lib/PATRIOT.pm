#!/usr/bin/perl
package PATRIOT;
use strict;
use Data::Dumper;
use CGI;
use Term::ANSIColor;

my $DEBUG = 0;

# Work follow: 
#   1. login via QR code at page: https://pc.xuexi.cn/points/login.html
#   2. Get the cookie from chrome

sub login
{
    my ($self, $url, $user, $password) = @_;


}


sub LOG
{
    my ($self, $level, $message) = @_;    
    my $now = &getTime();   # get current time 

    ### define the colr ###
    my $INFO_Color = 'green';
    my $WARN_Color = 'yellow';
    my $DEBUG_Color = 'bright_cyan';
    my $ERROR_Color = 'bright_red';
    my $FATAL_Color = 'red';

    if ($level =~ /INFO/i)
    {
        &printColor('self',$INFO_Color,"$now [INFO] $message\n");
    }
    elsif ($level =~ /WARN/i)
    {
        &printColor('self',$WARN_Color,"$now [WARN] $message\n");
    }
    elsif ($level =~ /DEBUG/i)
    {
        &printColor('self',$DEBUG_Color,"$now [DEBUG] $message\n");
    }
    elsif ($level =~ /ERROR/i)
    {
        &printColor('self',$ERROR_Color,"$now [ERROR] $message\n");
    }
    elsif ($level =~ /FATAL/i)
    {
        &printColor('self',$FATAL_Color,"$now [FATAL] $message\n");
        die "Script captured FATAL error, exit now!"
    }
    else
    {
        # $printColor('self',$WARN_Color,"$now [WARN] Unknown log level [$level]\n");
        print "$now $message\n"
    }
}

# get current time, with format yyyy-mm-dd hh:mm:ss
sub getTime
{
    my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime(time);
    $year += 1900; $mon += 1; 
    
    $sec = "0$sec" if ($sec < 10);
    $min = "0$min" if ($min < 10);
    $hour = "0hour" if ($hour < 10);

    my $now = "$year-$mon-$mday $hour:$min:$sec";
    return $now;
}

sub printColor
{
    my ($self,$Color,$MSG) = @_;
    print color "$Color"; print "$MSG"; print color 'reset';
}

1; ## must return true for PM file