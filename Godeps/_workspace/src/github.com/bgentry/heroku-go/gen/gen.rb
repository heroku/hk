#!/usr/bin/env ruby

require 'erubis'
require 'multi_json'

RESOURCE_TEMPLATE = <<-RESOURCE_TEMPLATE
// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

<%- if schemas[key]['properties'] && schemas[key]['properties'].any?{|p, v| resolve_typedef(v).end_with?("time.Time") } %>
import (
	"time"
)
<%- end %>

<%- if definition['properties'] %>
  <%- description = markdown_free(definition["description"] || "") %>
  <%- word_wrap(description, line_width: 77).split("\n").each do |line| %>
    // <%= line %>
  <%- end %>
  type <%= resource_class %> struct {
  <%- definition['properties'].each do |propname, propdef| %>
    <%- resolved_propdef = resolve_propdef(propdef) %>
    // <%= resolved_propdef["description"] %>
    <%- type = resolve_typedef(resolved_propdef) %>
    <%- if type =~ /\\*?struct/ %>
      <%= titlecase(propname) %> <%= type %> {
        <%- resolved_propdef["properties"].each do |subpropname, subpropdef| %>
          <%= titlecase(subpropname) %> <%= resolve_typedef(subpropdef) %> `json:"<%= subpropname %>"`
        <%- end %>
      } `json:"<%= propname %>"`
    <%- else %>
      <%= titlecase(propname) %> <%= resolve_typedef(propdef) %> `json:"<%= propname %>"`
    <%- end %>

  <%- end %>
  }
<%- end %>

<%- definition["links"].each do |link| %>
  <%- func_name = titlecase(key.downcase. + "-" + link["title"]) %>
  <%- func_args = [] %>
  <%- func_args << (variablecase(parent_resource_instance) + 'Identity string') if parent_resource_instance %>
  <%- func_args += func_args_from_model_and_link(definition, key, link) %>
  <%- return_values = return_values_from_link(key, link) %>
  <%- path = link['href'].gsub("{(%2Fschema%2F\#{key}%23%2Fdefinitions%2Fidentity)}", '"+' + variablecase(resource_instance) + 'Identity') %>
  <%- if parent_resource_instance %>
    <%- path = path.gsub("{(%2Fschema%2F" + parent_resource_instance + "%23%2Fdefinitions%2Fidentity)}", '" + ' + variablecase(parent_resource_instance) + 'Identity + "') %>
  <%- end %>
  <%- path = ensure_balanced_end_quote(ensure_open_quote(path)) %>

  <%- word_wrap(markdown_free(link["description"]), line_width: 77).split("\n").each do |line| %>
    // <%= line %>
  <%- end %>
  <%- func_arg_comments = [] %>
  <%- func_arg_comments << (variablecase(parent_resource_instance) + "Identity is the unique identifier of the " + key + "'s " + parent_resource_instance + ".") if parent_resource_instance %>
  <%- func_arg_comments += func_arg_comments_from_model_and_link(definition, key, link) %>
  //
  <%- word_wrap(func_arg_comments.join(" "), line_width: 77).split("\n").each do |comment| %>
    // <%= comment %>
  <%- end %>
  <%- flat_postval = link["schema"] && link["schema"]["additionalProperties"] == false %>
  <%- required = (link["schema"] && link["schema"]["required"]) || [] %>
  <%- optional = ((link["schema"] && link["schema"]["properties"]) || {}).keys - required %>
  <%- postval = if flat_postval %>
    <%-           "options" %>
    <%-         elsif required.empty? && optional.empty? %>
    <%-           "nil" %>
    <%-         elsif required.empty? %>
    <%-           "options" %>
    <%-         else %>
    <%-           "params" %>
    <%-         end %>
  <%- hasCustomType = !schemas[key]["properties"].nil? %>
  func (c *Client) <%= func_name + "(" + func_args.join(', ') %>) <%= return_values %> {
    <%- case link["rel"] %>
    <%- when "create" %>
      <%- if !required.empty? %>
        <%= Erubis::Eruby.new(LINK_PARAMS_TEMPLATE).result({modelname: key, link: link, required: required, optional: optional}).strip %>
      <%- end %>
      var <%= variablecase(key + '-res') %> <%= titlecase(key) %>
      return &<%= variablecase(key + '-res') %>, c.Post(&<%= variablecase(key + '-res') %>, <%= path %>, <%= postval %>)
    <%- when "self" %>
      var <%= variablecase(key) %> <%= hasCustomType ? titlecase(key) : "map[string]string" %>
      return <%= "&" if hasCustomType%><%= variablecase(key) %>, c.Get(&<%= variablecase(key) %>, <%= path %>)
    <%- when "destroy" %>
      return c.Delete(<%= path %>)
    <%- when "update" %>
      <%- if !required.empty? %>
        <%= Erubis::Eruby.new(LINK_PARAMS_TEMPLATE).result({modelname: key, link: link, required: required, optional: optional}).strip %>
      <%- end %>
      <%- if link["title"].include?("Batch") %>
        var <%= variablecase(key + 's-res') %> []<%= titlecase(key) %>
        return <%= variablecase(key + 's-res') %>, c.Patch(&<%= variablecase(key + 's-res') %>, <%= path %>, <%= postval %>)
      <%- else %>
        var <%= variablecase(key + '-res') %> <%= hasCustomType ? titlecase(key) : "map[string]string" %>
        return <%= "&" if hasCustomType%><%= variablecase(key + '-res') %>, c.Patch(&<%= variablecase(key + '-res') %>, <%= path %>, <%= postval %>)
      <%- end %>
    <%- when "instances" %>
      req, err := c.NewRequest("GET", <%= path %>, nil)
      if err != nil {
        return nil, err
      }

      if lr != nil {
        lr.SetHeader(req)
      }

      var <%= variablecase(key + 's-res') %> []<%= titlecase(key) %>
      return <%= variablecase(key + 's-res') %>, c.DoReq(req, &<%= variablecase(key + 's-res') %>)
    <%- end %>
  }

  <%- if %w{create update}.include?(link["rel"]) && link["schema"] && link["schema"]["properties"] %>
    <%- if !required.empty? %>
      <%- structs = required.select {|p| resolve_typedef(link["schema"]["properties"][p]) == "struct" } %>
      <%- structs.each do |propname| %>
        <%- typename = titlecase([key, link["title"], propname].join("-")) %>
        // <%= typename %> used in <%= func_name %> as the <%= definition["properties"][propname]["description"] %>
        type <%= typename %> struct {
          <%- link["schema"]["properties"][propname]["properties"].each do |subpropname, subval| %>
            <%- propdef = definition["properties"][propname]["properties"][subpropname] %>
            <%- description = resolve_propdef(propdef)["description"] %>
            <%- word_wrap(description, line_width: 77).split("\n").each do |line| %>
              // <%= line %>
            <%- end %>
            <%= titlecase(subpropname) %> <%= resolve_typedef(subval) %> `json:"<%= subpropname %>"`

          <%- end %>
        }
      <%- end %>
      <%- arr_structs = required.select {|p| resolve_typedef(link["schema"]["properties"][p]) == "[]struct" } %>
      <%- arr_structs.each do |propname| %>
        <%- # special case for arrays of structs (like FormationBulkUpdate) %>
        <%- typename = titlecase([key, link["title"], "opts"].join("-")) %>
        <%- typedef = resolve_propdef(link["schema"]["properties"][propname]["items"]) %>

        type <%= typename %> struct {
          <%- typedef["properties"].each do |subpropname, subref| %>
            <%- propdef = resolve_propdef(subref) %>
            <%- description = resolve_propdef(propdef)["description"] %>
            <%- is_required = typedef["required"].include?(subpropname) %>
            <%- word_wrap(description, line_width: 77).split("\n").each do |line| %>
              // <%= line %>
            <%- end %>
            <%= titlecase(subpropname) %> <%= "*" unless is_required %><%= resolve_typedef(propdef) %> `json:"<%= subpropname %><%= ",omitempty" unless is_required %>"`

          <%- end %>
        }
      <%- end %>
    <%- end %>
    <%- if !optional.empty? %>
      // <%= func_name %>Opts holds the optional parameters for <%= func_name %>
      type <%= func_name %>Opts struct {
        <%- optional.each do |propname| %>
          <%- if definition['properties'][propname] && definition['properties'][propname]['description'] %>
            // <%= definition['properties'][propname]['description'] %>
          <%- elsif definition["definitions"][propname] %>
            // <%= definition["definitions"][propname]["description"] %>
          <%- elsif link["schema"]["properties"][propname]["$ref"] %>
            // <%= resolve_propdef(link["schema"]["properties"][propname])["description"] %>
          <%- else %>
            // <%= link["schema"]["properties"][propname]["description"] %>
          <%- end %>
          <%= titlecase(propname) %> <%= type_for_link_opts_field(link, propname) %> `json:"<%= propname %>,omitempty"`
        <%- end %>
      }
    <%- end %>
  <%- end %>

<%- end %>
RESOURCE_TEMPLATE

LINK_PARAMS_TEMPLATE = <<-LINK_PARAMS_TEMPLATE
params := struct {
<%- required.each do |propname| %>
  <%- type = resolve_typedef(link["schema"]["properties"][propname]) %>
  <%- if type == "[]struct" %>
    <%- type = type.gsub("struct", titlecase([modelname, link["title"], "opts"].join("-"))) %>
  <%- elsif type == "struct" %>
    <%- type = titlecase([modelname, link["title"], propname].join("-")) %>
  <%- end %>
  <%= titlecase(propname) %> <%= type %> `json:"<%= propname %>"`
<%- end %>
<%- optional.each do |propname| %>
  <%= titlecase(propname) %> <%= type_for_link_opts_field(link, propname) %> `json:"<%= propname %>,omitempty"`
<%- end %>
}{
<%- required.each do |propname| %>
  <%= titlecase(propname) %>: <%= variablecase(propname) %>,
<%- end %>
<%- optional.each do |propname| %>
  <%= titlecase(propname) %>: options.<%= titlecase(propname) %>,
<%- end %>
}
LINK_PARAMS_TEMPLATE

#   definition:               data,
#   key:                      modelname,
#   parent_resource_class:    parent_resource_class,
#   parent_resource_identity: parent_resource_identity,
#   parent_resource_instance: parent_resource_instance,
#   resource_class:           resource_class,
#   resource_instance:        resource_instance,
#   resource_proxy_class:     resource_proxy_class,
#   resource_proxy_instance:  resource_proxy_instance

module Generator
  extend self

  def ensure_open_quote(str)
    str[0] == '"' ? str : "\"#{str}"
  end

  def ensure_balanced_end_quote(str)
    (str.count('"') % 2) == 1 ? "#{str}\"" : str
  end

  def must_end_with(str, ending)
    str.end_with?(ending) ? str : "#{str}#{ending}"
  end

  def word_wrap(text, options = {})
    line_width = options.fetch(:line_width, 80)

    text.split("\n").collect do |line|
      line.length > line_width ? line.gsub(/(.{1,#{line_width}})(\s+|$)/, "\\1\n").strip : line
    end * "\n"
  end

  def markdown_free(text)
    text.gsub(/\[(?<linktext>[^\]]*)\](?<linkurl>\(.*\))/, '\k<linktext>').
      gsub(/`(?<rawtext>[^\]]*)`/, '\k<rawtext>').gsub("NULL", "nil")
  end

  def variablecase(str)
    words = str.gsub('_','-').gsub(' ','-').split('-')
    (words[0...1] + words[1..-1].map {|k| k[0...1].upcase + k[1..-1]}).join
  end

  def titlecase(str)
    str.gsub('_','-').gsub(' ','-').split('-').map do |k|
      if k.downcase == "url" # special case so Url becomes URL
        k.upcase
      elsif k.downcase == "oauth" # special case so Oauth becomes OAuth
        "OAuth"
      else
        k[0...1].upcase + k[1..-1]
      end
    end.join
  end

  def resolve_typedef(propdef)
    if types = propdef["type"]
      null = types.include?("null")
      tname = case (types - ["null"]).first
              when "boolean"
                "bool"
              when "integer"
                "int"
              when "string"
                format = propdef["format"]
                format && format == "date-time" ? "time.Time" : "string"
              when "object"
                if propdef["additionalProperties"] == false
                  if propdef["patternProperties"]
                    "map[string]string"
                  else
                    # special case for arrays of structs (like FormationBulkUpdate)
                    "struct"
                  end
                else
                  "struct"
                end
              when "array"
                arraytype = if propdef["items"]["$ref"]
                  resolve_typedef(propdef["items"])
                else
                  propdef["items"]["type"]
                end
                "[]#{arraytype}"
              else
                types.first
              end
      null ? "*#{tname}" : tname
    elsif propdef["anyOf"]
      # identity cross-reference, cheat because these are always strings atm
      "string"
    elsif propdef["additionalProperties"] == false
      # inline object
      propdef
    elsif ref = propdef["$ref"]
      matches = ref.match(/\/schema\/([\w-]+)#\/definitions\/([\w-]+)/)
      schemaname, fieldname = matches[1..2]
      resolve_typedef(schemas[schemaname]["definitions"][fieldname])
    else
      raise "WTF #{propdef}"
    end
  end

  def type_for_link_opts_field(link, propname, nullable = true)
    resulttype = resolve_typedef(link["schema"]["properties"][propname])
    if nullable && !resulttype.start_with?("*")
      resulttype = "*#{resulttype}"
    elsif !nullable
      resulttype = resulttype.gsub("*", "")
    end
    resulttype
  end

  def type_from_types_and_format(types, format)
    case types.first
    when "boolean"
      "bool"
    when "integer"
      "int"
    when "string"
      format && format == "date-time" ? "time.Time" : "string"
    else
      types.first
    end
  end

  def return_values_from_link(modelname, link)
    if !schemas[modelname]["properties"]
      # structless type like ConfigVar
      "(map[string]string, error)"
    else
      case link["rel"]
      when "destroy"
        "error"
      when "instances"
        "([]#{titlecase(modelname)}, error)"
      else
        if link["title"].include?("Batch")
          "([]#{titlecase(modelname)}, error)"
        else
          "(*#{titlecase(modelname)}, error)"
        end
      end
    end
  end

  def func_args_from_model_and_link(definition, modelname, link)
    args = []
    required = (link["schema"] && link["schema"]["required"]) || []
    optional = ((link["schema"] && link["schema"]["properties"]) || {}).keys - required

    # check if this link's href requires the model's identity
    match = link["href"].match(%r{%2Fschema%2F#{modelname}%23%2Fdefinitions%2Fidentity})
    if %w{update destroy self}.include?(link["rel"]) && match
      args << "#{variablecase(modelname)}Identity string"
    end

    if %w{create update}.include?(link["rel"])
      if link["schema"]["additionalProperties"] == false
        # handle ConfigVar update
        args << "options map[string]*string"
      else
        required.each do |propname|
          type = type_for_link_opts_field(link, propname, false)
          if type == "[]struct"
            type = type.gsub("struct", titlecase([modelname, link["title"], "Opts"].join("-")))
          elsif type == "struct"
            type = type.gsub("struct", titlecase([modelname, link["title"], propname].join("-")))
          end
          args << "#{variablecase(propname)} #{type}"
        end
      end
      args << "options *#{titlecase(modelname)}#{link["rel"].capitalize}Opts" unless optional.empty?
    end

    if "instances" == link["rel"]
      args << "lr *ListRange"
    end

    args
  end

  def resolve_propdef(propdef)
    resolve_all_propdefs(propdef).first
  end

  def resolve_all_propdefs(propdef)
    if propdef["description"]
      [propdef]
    elsif ref = propdef["$ref"]
      matches = ref.match(/\/schema\/([\w-]+)#\/definitions\/([\w-]+)/)
      schemaname, fieldname = matches[1..2]
      resolve_all_propdefs(schemas[schemaname]["definitions"][fieldname])
    elsif anyof = propdef["anyOf"]
      # Identity
      anyof.map do |refhash|
        matches = refhash["$ref"].match(/\/schema\/([\w-]+)#\/definitions\/([\w-]+)/)
        schemaname, fieldname = matches[1..2]
        resolve_all_propdefs(schemas[schemaname]["definitions"][fieldname])
      end.flatten
    elsif propdef["type"] && propdef["type"].is_a?(Array) && propdef["type"].first == "object"
      # special case for params which are nested objects, like oauth-grant
      [propdef]
    else
      raise "WTF #{propdef}"
    end
  end

  def func_arg_comments_from_model_and_link(definition, modelname, link)
    args = []
    flat_postval = link["schema"] && link["schema"]["additionalProperties"] == false
    properties = (link["schema"] && link["schema"]["properties"]) || {}
    required_keys = (link["schema"] && link["schema"]["required"]) || []
    optional_keys = properties.keys - required_keys

    if %w{update destroy self}.include?(link["rel"])
      if flat_postval
        # special case for ConfigVar update w/ flat param struct
        desc = markdown_free(link["schema"]["description"])
        args << "options is the #{desc}."
      else
        args << "#{variablecase(modelname)}Identity is the unique identifier of the #{titlecase(modelname)}."
      end
    end

    if %w{create update}.include?(link["rel"])
      required_keys.each do |propname|
        rpresults = resolve_all_propdefs(link["schema"]["properties"][propname])
        if rpresults.size == 1
          if rpresults.first["properties"]
            # special case for things like OAuthToken with nested objects
            rpresults = resolve_all_propdefs(definition["properties"][propname])
          end
          args << "#{variablecase(propname)} is the #{must_end_with(rpresults.first["description"] || "", ".")}"
        elsif rpresults.size == 2
          args << "#{variablecase(propname)} is the #{rpresults.first["description"]} or #{must_end_with(rpresults.last["description"] || "", ".")}"
        else
          raise "Didn't expect 3 rpresults"
        end
      end
      args << "options is the struct of optional parameters for this action." unless optional_keys.empty?
    end

    if "instances" == link["rel"]
      args << "lr is an optional ListRange that sets the Range options for the paginated list of results."
    end

    case link["rel"]
    when "create"
      ["options is the struct of optional parameters for this action."]
    when "update"
      ["#{variablecase(modelname)}Identity is the unique identifier of the #{titlecase(modelname)}.",
       "options is the struct of optional parameters for this action."]
    when "destroy", "self"
      ["#{variablecase(modelname)}Identity is the unique identifier of the #{titlecase(modelname)}."]
    when "instances"
      ["lr is an optional ListRange that sets the Range options for the paginated list of results."]
    else
      []
    end
    args
  end

  def resource_instance_from_model(modelname)
    modelname.downcase.split('-').join('_')
  end

  def schemas
    @@schemas ||= {}
  end

  def load_model_schema(modelname)
    schema_path = File.expand_path("./schema/#{modelname}.json")
    schemas[modelname] = MultiJson.load(File.read(schema_path))
  end

  def generate_model(modelname)
    if !schemas[modelname]
      puts "no schema for #{modelname}" && return
    end
    if schemas[modelname]['links'].empty?
      puts "no links for #{modelname}"
    end

    resource_class = titlecase(modelname)
    resource_instance = resource_instance_from_model(modelname)

    resource_proxy_class = resource_class + 's'
    resource_proxy_instance = resource_instance + 's'

    parent_resource_class, parent_resource_identity, parent_resource_instance = if schemas[modelname]['links'].all? {|link| link['href'].include?('{(%2Fschema%2Fapp%23%2Fdefinitions%2Fidentity)}')}
      ['App', 'app_identity', 'app']
    end

    data = Erubis::Eruby.new(RESOURCE_TEMPLATE).result({
      definition:               schemas[modelname],
      key:                      modelname,
      parent_resource_class:    parent_resource_class,
      parent_resource_identity: parent_resource_identity,
      parent_resource_instance: parent_resource_instance,
      resource_class:           resource_class,
      resource_instance:        resource_instance,
      resource_proxy_class:     resource_proxy_class,
      resource_proxy_instance:  resource_proxy_instance
    })

    path = File.expand_path(File.join(File.dirname(__FILE__), '..', "#{modelname.gsub('-', '_')}.go"))
    File.open(path, 'w') do |file|
      file.write(data)
    end
    %x( go fmt #{path} )
  end
end

include Generator

models = Dir.glob("schema/*.json").map{|f| f.gsub(".json", "") }.map{|f| f.gsub("schema/", "")}

models.each do |modelname|
  puts "Loading #{modelname}..."
  Generator.load_model_schema(modelname)
end

models.each do |modelname|
  puts "Generating #{modelname}..."
  if (Generator.schemas[modelname]["links"] || []).empty? && Generator.schemas[modelname]["properties"].empty?
    puts "-- skipping #{modelname} because it has no links or properties"
  else
    Generator.generate_model(modelname)
  end
end
