Pod::Spec.new do |spec|
  spec.name         = 'Gdos'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/doslink/dos'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Doslink Client'
  spec.source       = { :git => 'https://github.com/doslink/dos.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gdos.framework'

	spec.prepare_command = <<-CMD
    curl https://gdosstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gdos.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
