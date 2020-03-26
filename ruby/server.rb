# server.rb

# https://spin.atomicobject.com/2013/11/12/production-logging-sinatra/

require 'sinatra'
require "sinatra/namespace"
require 'sinatra/base'

require 'logger'

require "json"
require "addressable/uri"
require "mysql2"

sqlclient = Mysql2::Client.new(:host => "192.168.89.33", :username => "eliezer", :password => "password", :database => "yt" , :reconnect => true )


def youtube_id(youtube_url)
    regex = /(?:youtube(?:-nocookie)?\.com\/(?:[^\/\n\s]+\/\S+\/|(?:v|e(?:mbed)?)\/|\S*?[?&]v=)|youtu\.be\/)([a-zA-Z0-9_-]{11})(?:\&|\?|$)/
    match = regex.match(youtube_url)
    if match && !match[1].empty?
      match[1]
    else
      nil
    end
end

list = { 
	"cySa6WHjUUA" => { "categories" => [ "nudity" ] },
	"N5YPPU84PNk" => { "categories" => [ "nudity" ] },


}

class SinatraLoggingExample < Sinatra::Base

#  configure :production do
#    set :haml, { :ugly=>true }
#    set :clean_trace, true
#
#    Dir.mkdir('logs') unless File.exist?('logs')
#
#    $logger = Logger.new('logs/common.log','weekly')
#    $logger.level = Logger::WARN
#
#    # Spit stdout and stderr to a file during production
#    # in case something goes wrong
#    $stdout.reopen("logs/output.log", "w")
##    $stdout.sync = true
#
#    $stderr.reopen($stdout)
##    $stderr.sync = true
#
#  end
#
#  configure :development do
#    $logger = Logger.new(STDOUT)
#  end


before do
	content_type 'application/json'
end

get '/yt/' do
	  res = {}
	  if params[:url] != nil 
		  query_url = params[:url] 
		  begin
		  query_uri = Addressable::URI.parse(query_url)
		  rescue => e
			  STDERR.puts(e)
			  STDERR.puts(e.inspect)
			  res["error"] = 1
			  res["msg"] = e.inspect
			  return JSON.pretty_generate(res)+"\n"
		  end
		  id = youtube_id(query_url)
		  if id != nil
			  escaped_input = sqlclient.escape(id)
			  results = sqlclient.query("SELECT * FROM videos WHERE videoid='#{escaped_input}'")
			  if results.size == 1
				  puts results[0]
				  rate = results[0]["rated"] 
				  if rate > 60
					  res["nudity"] = 1000
				  end
			  end
		  end
	  end
     return JSON.pretty_generate(res)+"\n"
end

end
